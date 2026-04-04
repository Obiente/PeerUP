package settings

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"connectrpc.com/connect"
	"go.uber.org/zap"

	settingsv1 "github.com/Obiente/Uppe/apps/server/gen/settings/v1"
	"github.com/Obiente/Uppe/apps/server/gen/settings/v1/settingsv1connect"
	"github.com/Obiente/Uppe/apps/server/internal/db"
)

type SettingsService struct {
	logger   *zap.Logger
	database db.Database
}

var _ settingsv1connect.SettingsServiceHandler = (*SettingsService)(nil)

func NewSettingsService(logger *zap.Logger, database db.Database) *SettingsService {
	return &SettingsService{
		logger:   logger,
		database: database,
	}
}

func (s *SettingsService) GetSettings(
	ctx context.Context,
	_ *connect.Request[settingsv1.GetSettingsRequest],
) (*connect.Response[settingsv1.Settings], error) {
	settings, err := s.loadSettings(ctx)
	if err != nil {
		s.logger.Error("Failed to load settings", zap.Error(err))
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to load settings: %w", err))
	}

	return connect.NewResponse(settings), nil
}

func (s *SettingsService) UpdateSettings(
	ctx context.Context,
	req *connect.Request[settingsv1.UpdateSettingsRequest],
) (*connect.Response[settingsv1.Settings], error) {
	if req.Msg.Settings == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("settings payload is required"))
	}

	if err := s.storeSettings(ctx, req.Msg.Settings); err != nil {
		s.logger.Error("Failed to update settings", zap.Error(err))
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update settings: %w", err))
	}

	settings, err := s.loadSettings(ctx)
	if err != nil {
		s.logger.Error("Failed to reload settings", zap.Error(err))
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to reload settings: %w", err))
	}

	return connect.NewResponse(settings), nil
}

func (s *SettingsService) GetNodeIdentity(
	ctx context.Context,
	_ *connect.Request[settingsv1.GetNodeIdentityRequest],
) (*connect.Response[settingsv1.NodeIdentity], error) {
	identity, err := s.database.GetNodeIdentity(ctx)
	if err != nil {
		s.logger.Error("Failed to load node identity", zap.Error(err))
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to load node identity: %w", err))
	}

	return connect.NewResponse(&settingsv1.NodeIdentity{
		NodeId:               identity.NodeID,
		NodeName:             identity.NodeName,
		JoinedNetworkAt:      identity.JoinedNetworkAt.Unix(),
		ContributionScore:    identity.ContributionScore,
		TotalChecksPerformed: identity.TotalChecksPerformed,
		TotalChecksReceived:  identity.TotalChecksReceived,
	}), nil
}

func (s *SettingsService) loadSettings(ctx context.Context) (*settingsv1.Settings, error) {
	getString := func(key, def string) (string, error) {
		v, ok, err := s.database.GetSetting(ctx, key)
		if err != nil {
			return "", err
		}
		if !ok || v == "" {
			return def, nil
		}
		return v, nil
	}

	getBool := func(key string, def bool) (bool, error) {
		v, ok, err := s.database.GetSetting(ctx, key)
		if err != nil {
			return false, err
		}
		if !ok || v == "" {
			return def, nil
		}
		b, parseErr := strconv.ParseBool(v)
		if parseErr != nil {
			return def, nil
		}
		return b, nil
	}

	getInt32 := func(key string, def int32) (int32, error) {
		v, ok, err := s.database.GetSetting(ctx, key)
		if err != nil {
			return 0, err
		}
		if !ok || v == "" {
			return def, nil
		}
		n, parseErr := strconv.Atoi(v)
		if parseErr != nil {
			return def, nil
		}
		return int32(n), nil
	}

	getInt64 := func(key string, def int64) (int64, error) {
		v, ok, err := s.database.GetSetting(ctx, key)
		if err != nil {
			return 0, err
		}
		if !ok || v == "" {
			return def, nil
		}
		n, parseErr := strconv.ParseInt(v, 10, 64)
		if parseErr != nil {
			return def, nil
		}
		return n, nil
	}

	getFloat64 := func(key string, def float64) (float64, error) {
		v, ok, err := s.database.GetSetting(ctx, key)
		if err != nil {
			return 0, err
		}
		if !ok || v == "" {
			return def, nil
		}
		n, parseErr := strconv.ParseFloat(v, 64)
		if parseErr != nil {
			return def, nil
		}
		return n, nil
	}

	getStringArray := func(key string) ([]string, error) {
		v, ok, err := s.database.GetSetting(ctx, key)
		if err != nil {
			return nil, err
		}
		if !ok || v == "" {
			return []string{}, nil
		}
		var values []string
		if err := json.Unmarshal([]byte(v), &values); err != nil {
			return []string{}, nil
		}
		return values, nil
	}

	displayName, err := getString("display_name", "Uppe. Node")
	if err != nil {
		return nil, err
	}
	email, err := getString("email", "")
	if err != nil {
		return nil, err
	}
	timezone, err := getString("timezone", "UTC")
	if err != nil {
		return nil, err
	}

	emailNotifications, err := getBool("email_notifications", true)
	if err != nil {
		return nil, err
	}
	incidentReports, err := getBool("incident_reports", false)
	if err != nil {
		return nil, err
	}
	networkUpdates, err := getBool("network_updates", true)
	if err != nil {
		return nil, err
	}

	discoveryMethodRaw, err := getString("discovery_method", "dht_and_bootstrap")
	if err != nil {
		return nil, err
	}
	contributeToNetwork, err := getBool("contribute_to_network", true)
	if err != nil {
		return nil, err
	}
	maxBandwidth, err := getInt64("max_bandwidth_mb_per_day", 100)
	if err != nil {
		return nil, err
	}
	maxConcurrentChecks, err := getInt32("max_concurrent_checks", 50)
	if err != nil {
		return nil, err
	}
	autoAcceptRequests, err := getBool("auto_accept_requests", true)
	if err != nil {
		return nil, err
	}
	bootstrapNodes, err := getStringArray("bootstrap_nodes")
	if err != nil {
		return nil, err
	}
	manualPeers, err := getStringArray("manual_peers")
	if err != nil {
		return nil, err
	}

	defaultInterval, err := getInt32("default_interval_seconds", 60)
	if err != nil {
		return nil, err
	}
	defaultTimeout, err := getInt32("default_timeout_seconds", 10)
	if err != nil {
		return nil, err
	}
	detailedLogging, err := getBool("detailed_logging", false)
	if err != nil {
		return nil, err
	}
	retentionDays, err := getInt32("result_retention_days", 30)
	if err != nil {
		return nil, err
	}

	clusterName, err := getString("cluster_name", "My Cluster")
	if err != nil {
		return nil, err
	}
	clusterPublic, err := getBool("cluster_is_public", true)
	if err != nil {
		return nil, err
	}
	maxClusterSize, err := getInt32("cluster_max_size", 25)
	if err != nil {
		return nil, err
	}
	joinPolicyRaw, err := getString("cluster_join_policy", "open")
	if err != nil {
		return nil, err
	}
	minContributionScore, err := getFloat64("cluster_min_contribution_score", 1.5)
	if err != nil {
		return nil, err
	}

	return &settingsv1.Settings{
		Account: &settingsv1.AccountSettings{
			DisplayName: displayName,
			Email:       email,
			Timezone:    timezone,
		},
		Notifications: &settingsv1.NotificationSettings{
			EmailNotifications:           emailNotifications,
			IncidentReports:              incidentReports,
			NetworkUpdates:               networkUpdates,
			DowntimeAlertThresholdSeconds: 300,
			LatencyAlertThresholdMs:      1000,
		},
		Network: &settingsv1.NetworkSettings{
			DiscoveryMethod:   mapDiscoveryMethod(discoveryMethodRaw),
			ContributeToNetwork: contributeToNetwork,
			MaxBandwidthMbPerDay: maxBandwidth,
			MaxConcurrentChecks: maxConcurrentChecks,
			AutoAcceptRequests: autoAcceptRequests,
			BootstrapNodes:     bootstrapNodes,
			ManualPeers:        manualPeers,
		},
		Monitoring: &settingsv1.MonitoringSettings{
			DefaultIntervalSeconds: defaultInterval,
			DefaultTimeoutSeconds:  defaultTimeout,
			DetailedLogging:        detailedLogging,
			ResultRetentionDays:    retentionDays,
		},
		Cluster: &settingsv1.ClusterSettings{
			ClusterName:          clusterName,
			IsPublic:             clusterPublic,
			MaxClusterSize:       maxClusterSize,
			JoinPolicy:           mapJoinPolicy(joinPolicyRaw),
			MinContributionScore: minContributionScore,
		},
	}, nil
}

func (s *SettingsService) storeSettings(ctx context.Context, settings *settingsv1.Settings) error {
	if settings.Account != nil {
		if err := s.database.SetSetting(ctx, "display_name", settings.Account.DisplayName); err != nil {
			return err
		}
		if err := s.database.SetSetting(ctx, "email", settings.Account.Email); err != nil {
			return err
		}
		if err := s.database.SetSetting(ctx, "timezone", settings.Account.Timezone); err != nil {
			return err
		}
	}

	if settings.Notifications != nil {
		if err := s.database.SetSetting(ctx, "email_notifications", strconv.FormatBool(settings.Notifications.EmailNotifications)); err != nil {
			return err
		}
		if err := s.database.SetSetting(ctx, "incident_reports", strconv.FormatBool(settings.Notifications.IncidentReports)); err != nil {
			return err
		}
		if err := s.database.SetSetting(ctx, "network_updates", strconv.FormatBool(settings.Notifications.NetworkUpdates)); err != nil {
			return err
		}
	}

	if settings.Network != nil {
		if err := s.database.SetSetting(ctx, "discovery_method", unmapDiscoveryMethod(settings.Network.DiscoveryMethod)); err != nil {
			return err
		}
		if err := s.database.SetSetting(ctx, "contribute_to_network", strconv.FormatBool(settings.Network.ContributeToNetwork)); err != nil {
			return err
		}
		if err := s.database.SetSetting(ctx, "max_bandwidth_mb_per_day", strconv.FormatInt(settings.Network.MaxBandwidthMbPerDay, 10)); err != nil {
			return err
		}
		if err := s.database.SetSetting(ctx, "max_concurrent_checks", strconv.FormatInt(int64(settings.Network.MaxConcurrentChecks), 10)); err != nil {
			return err
		}
		if err := s.database.SetSetting(ctx, "auto_accept_requests", strconv.FormatBool(settings.Network.AutoAcceptRequests)); err != nil {
			return err
		}
		if err := setStringArray(ctx, s.database, "bootstrap_nodes", settings.Network.BootstrapNodes); err != nil {
			return err
		}
		if err := setStringArray(ctx, s.database, "manual_peers", settings.Network.ManualPeers); err != nil {
			return err
		}
	}

	if settings.Monitoring != nil {
		if err := s.database.SetSetting(ctx, "default_interval_seconds", strconv.FormatInt(int64(settings.Monitoring.DefaultIntervalSeconds), 10)); err != nil {
			return err
		}
		if err := s.database.SetSetting(ctx, "default_timeout_seconds", strconv.FormatInt(int64(settings.Monitoring.DefaultTimeoutSeconds), 10)); err != nil {
			return err
		}
		if err := s.database.SetSetting(ctx, "detailed_logging", strconv.FormatBool(settings.Monitoring.DetailedLogging)); err != nil {
			return err
		}
		if err := s.database.SetSetting(ctx, "result_retention_days", strconv.FormatInt(int64(settings.Monitoring.ResultRetentionDays), 10)); err != nil {
			return err
		}
	}

	if settings.Cluster != nil {
		if err := s.database.SetSetting(ctx, "cluster_name", settings.Cluster.ClusterName); err != nil {
			return err
		}
		if err := s.database.SetSetting(ctx, "cluster_is_public", strconv.FormatBool(settings.Cluster.IsPublic)); err != nil {
			return err
		}
		if err := s.database.SetSetting(ctx, "cluster_max_size", strconv.FormatInt(int64(settings.Cluster.MaxClusterSize), 10)); err != nil {
			return err
		}
		if err := s.database.SetSetting(ctx, "cluster_join_policy", unmapJoinPolicy(settings.Cluster.JoinPolicy)); err != nil {
			return err
		}
		if err := s.database.SetSetting(ctx, "cluster_min_contribution_score", strconv.FormatFloat(settings.Cluster.MinContributionScore, 'f', -1, 64)); err != nil {
			return err
		}
	}

	return nil
}

func setStringArray(ctx context.Context, database db.Database, key string, values []string) error {
	data, err := json.Marshal(values)
	if err != nil {
		return err
	}
	return database.SetSetting(ctx, key, string(data))
}

func mapDiscoveryMethod(v string) settingsv1.PeerDiscoveryMethod {
	switch v {
	case "dht_only":
		return settingsv1.PeerDiscoveryMethod_PEER_DISCOVERY_DHT_ONLY
	case "bootstrap_only":
		return settingsv1.PeerDiscoveryMethod_PEER_DISCOVERY_BOOTSTRAP_ONLY
	case "manual":
		return settingsv1.PeerDiscoveryMethod_PEER_DISCOVERY_MANUAL
	default:
		return settingsv1.PeerDiscoveryMethod_PEER_DISCOVERY_DHT_AND_BOOTSTRAP
	}
}

func unmapDiscoveryMethod(v settingsv1.PeerDiscoveryMethod) string {
	switch v {
	case settingsv1.PeerDiscoveryMethod_PEER_DISCOVERY_DHT_ONLY:
		return "dht_only"
	case settingsv1.PeerDiscoveryMethod_PEER_DISCOVERY_BOOTSTRAP_ONLY:
		return "bootstrap_only"
	case settingsv1.PeerDiscoveryMethod_PEER_DISCOVERY_MANUAL:
		return "manual"
	default:
		return "dht_and_bootstrap"
	}
}

func mapJoinPolicy(v string) settingsv1.ClusterJoinPolicy {
	switch v {
	case "approval":
		return settingsv1.ClusterJoinPolicy_CLUSTER_JOIN_POLICY_APPROVAL
	case "invite":
		return settingsv1.ClusterJoinPolicy_CLUSTER_JOIN_POLICY_INVITE
	default:
		return settingsv1.ClusterJoinPolicy_CLUSTER_JOIN_POLICY_OPEN
	}
}

func unmapJoinPolicy(v settingsv1.ClusterJoinPolicy) string {
	switch v {
	case settingsv1.ClusterJoinPolicy_CLUSTER_JOIN_POLICY_APPROVAL:
		return "approval"
	case settingsv1.ClusterJoinPolicy_CLUSTER_JOIN_POLICY_INVITE:
		return "invite"
	default:
		return "open"
	}
}
