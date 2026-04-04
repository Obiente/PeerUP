/**
 * Uppe. API Client
 * 
 * ConnectRPC client for communicating with the Uppe. API server
 * Supports Monitor and Result services
 */

import { createConnectTransport } from '@connectrpc/connect-web';
import { createPromiseClient } from '@connectrpc/connect';
import { MonitorService } from '../../gen/monitor/v1/monitor_connect';
import { NetworkService } from '../../gen/network/v1/network_connect';
import { ResultService } from '../../gen/result/v1/result_connect';
import { SettingsService } from '../../gen/settings/v1/settings_connect';
import { StatusPageService } from '../../gen/statuspage/v1/statuspage_connect';

// API configuration
const API_BASE_URL = import.meta.env.PUBLIC_API_URL || 'http://localhost:8080';

// Create transport
const transport = createConnectTransport({
  baseUrl: API_BASE_URL,
});

// Create promise-based clients
export const monitorClient = createPromiseClient(MonitorService, transport);
export const networkClient = createPromiseClient(NetworkService, transport);
export const resultClient = createPromiseClient(ResultService, transport);
export const settingsClient = createPromiseClient(SettingsService, transport);
export const statusPageClient = createPromiseClient(StatusPageService, transport);

/**
 * Health check - verify API is reachable
 */
export async function healthCheck(): Promise<boolean> {
  try {
    const response = await fetch(`${API_BASE_URL}/health`);
    return response.ok;
  } catch {
    return false;
  }
}
