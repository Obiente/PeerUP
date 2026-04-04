use std::collections::BTreeMap;
use std::time::{SystemTime, UNIX_EPOCH};

use anyhow::{Result, anyhow};
use ed25519_dalek::{Signature, Signer, Verifier, VerifyingKey};
use serde::Serialize;
use serde_json::Value;
use sha2::{Digest, Sha256};
use uuid::Uuid;

use crate::crypto::KeyPair;

#[derive(Debug, Clone, Serialize)]
struct SignableEventEnvelope<'a> {
    event_id: &'a str,
    event_type: &'a str,
    schema_version: i32,
    created_at: i64,
    actor_id: &'a str,
    actor_public_key_hex: String,
    resource_type: &'a str,
    resource_id: &'a str,
    parent_event_id: Option<String>,
    payload_json: &'a str,
    payload_hash: &'a str,
    capability_id: Option<&'a str>,
    delegated_by: Option<&'a str>,
    expires_at: Option<i64>,
    context_json: Option<&'a str>,
}

#[derive(Debug, Clone, serde::Serialize, serde::Deserialize)]
pub struct SignedEventEnvelope {
    pub event_id: Uuid,
    pub event_type: String,
    pub schema_version: i32,
    pub created_at: SystemTime,
    pub actor_id: String,
    pub actor_public_key: Vec<u8>,
    pub resource_type: String,
    pub resource_id: String,
    pub parent_event_id: Option<Uuid>,
    pub payload_json: String,
    pub payload_hash: String,
    pub capability_id: Option<String>,
    pub delegated_by: Option<String>,
    pub expires_at: Option<SystemTime>,
    pub context_json: Option<String>,
    pub signature: Vec<u8>,
}

#[derive(Debug, Clone)]
pub struct NewSignedEvent<'a, T: Serialize, C: Serialize> {
    pub event_type: &'a str,
    pub actor_id: &'a str,
    pub resource_type: &'a str,
    pub resource_id: &'a str,
    pub parent_event_id: Option<Uuid>,
    pub payload: &'a T,
    pub capability_id: Option<&'a str>,
    pub delegated_by: Option<&'a str>,
    pub expires_at: Option<SystemTime>,
    pub context: Option<&'a C>,
}

impl SignedEventEnvelope {
    pub const SCHEMA_VERSION: i32 = 1;

    pub fn sign<T: Serialize, C: Serialize>(
        spec: NewSignedEvent<'_, T, C>,
        keypair: &KeyPair,
    ) -> Result<Self> {
        let payload_json = canonical_json_string(&serde_json::to_value(spec.payload)?)?;
        let payload_hash = hex::encode(Sha256::digest(payload_json.as_bytes()));
        let context_json = match spec.context {
            Some(context) => Some(canonical_json_string(&serde_json::to_value(context)?)?),
            None => None,
        };

        let mut event = Self {
            event_id: Uuid::new_v4(),
            event_type: spec.event_type.to_string(),
            schema_version: Self::SCHEMA_VERSION,
            created_at: SystemTime::now(),
            actor_id: spec.actor_id.to_string(),
            actor_public_key: keypair.public_key_bytes().to_vec(),
            resource_type: spec.resource_type.to_string(),
            resource_id: spec.resource_id.to_string(),
            parent_event_id: spec.parent_event_id,
            payload_json,
            payload_hash,
            capability_id: spec.capability_id.map(ToOwned::to_owned),
            delegated_by: spec.delegated_by.map(ToOwned::to_owned),
            expires_at: spec.expires_at,
            context_json,
            signature: Vec::new(),
        };

        let signature = keypair.signing_key.sign(&event.signing_bytes()?);
        event.signature = signature.to_bytes().to_vec();
        Ok(event)
    }

    pub fn verify(&self) -> Result<bool> {
        let public_key_bytes: [u8; 32] = self
            .actor_public_key
            .as_slice()
            .try_into()
            .map_err(|_| anyhow!("invalid actor public key length"))?;
        let signature_bytes: [u8; 64] = self
            .signature
            .as_slice()
            .try_into()
            .map_err(|_| anyhow!("invalid signature length"))?;

        let expected_hash = hex::encode(Sha256::digest(self.payload_json.as_bytes()));
        if expected_hash != self.payload_hash {
            return Ok(false);
        }

        let verifying_key = VerifyingKey::from_bytes(&public_key_bytes)
            .map_err(|e| anyhow!("invalid actor public key: {e}"))?;
        let signature = Signature::from_bytes(&signature_bytes);
        Ok(verifying_key.verify(&self.signing_bytes()?, &signature).is_ok())
    }

    #[allow(dead_code)]
    pub fn payload_value(&self) -> Result<Value> {
        Ok(serde_json::from_str(&self.payload_json)?)
    }

    #[allow(dead_code)]
    pub fn context_value(&self) -> Result<Option<Value>> {
        self.context_json
            .as_ref()
            .map(|json| serde_json::from_str(json).map_err(Into::into))
            .transpose()
    }

    fn signing_bytes(&self) -> Result<Vec<u8>> {
        Ok(serde_json::to_vec(&SignableEventEnvelope {
            event_id: &self.event_id.to_string(),
            event_type: &self.event_type,
            schema_version: self.schema_version,
            created_at: system_time_to_i64(self.created_at),
            actor_id: &self.actor_id,
            actor_public_key_hex: hex::encode(&self.actor_public_key),
            resource_type: &self.resource_type,
            resource_id: &self.resource_id,
            parent_event_id: self.parent_event_id.map(|id| id.to_string()),
            payload_json: &self.payload_json,
            payload_hash: &self.payload_hash,
            capability_id: self.capability_id.as_deref(),
            delegated_by: self.delegated_by.as_deref(),
            expires_at: self.expires_at.map(system_time_to_i64),
            context_json: self.context_json.as_deref(),
        })?)
    }
}

pub fn canonical_json_string(value: &Value) -> Result<String> {
    Ok(serde_json::to_string(&canonicalize_value(value))?)
}

fn canonicalize_value(value: &Value) -> Value {
    match value {
        Value::Object(map) => {
            let mut ordered = BTreeMap::new();
            for (key, value) in map {
                ordered.insert(key.clone(), canonicalize_value(value));
            }
            serde_json::to_value(ordered).expect("ordered map should serialize")
        }
        Value::Array(values) => Value::Array(values.iter().map(canonicalize_value).collect()),
        _ => value.clone(),
    }
}

pub fn system_time_to_i64(time: SystemTime) -> i64 {
    time.duration_since(UNIX_EPOCH).unwrap_or_default().as_secs() as i64
}

#[allow(dead_code)]
pub fn i64_to_system_time(timestamp: i64) -> SystemTime {
    UNIX_EPOCH + std::time::Duration::from_secs(timestamp.max(0) as u64)
}

#[cfg(test)]
mod tests {
    use super::*;
    use peerup::crypto::generate_keypair;
    use serde_json::json;

    #[test]
    fn canonical_json_orders_keys_recursively() {
        let value = json!({
            "z": 1,
            "a": { "b": 2, "a": 1 }
        });
        assert_eq!(canonical_json_string(&value).unwrap(), r#"{"a":{"a":1,"b":2},"z":1}"#);
    }

    #[test]
    fn signed_event_round_trips_verification() {
        let keypair = generate_keypair();
        let payload = json!({ "status": "up", "monitor_id": "mon-1" });

        let event = SignedEventEnvelope::sign(
            NewSignedEvent::<_, serde_json::Value> {
                event_type: "monitor.result.recorded",
                actor_id: "peer-1",
                resource_type: "monitor",
                resource_id: "mon-1",
                parent_event_id: None,
                payload: &payload,
                capability_id: None,
                delegated_by: None,
                expires_at: None,
                context: None,
            },
            &keypair,
        )
        .unwrap();

        assert!(event.verify().unwrap());
    }
}
