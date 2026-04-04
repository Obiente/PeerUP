/**
 * Uppe. API Module
 * 
 * Direct ConnectRPC client access for fully static bundle compatibility.
 * All pages import and use these clients directly without wrapper functions,
 * enabling the frontend to be deployed as a static site in the future.
 * 
 * Usage:
 * ```astro
 * ---
 * import { monitorClient, resultClient } from '../lib/api';
 * 
 * const monitors = await monitorClient.listMonitors({ page: 1, pageSize: 100 });
 * const results = await resultClient.getResults({ monitorId: 'mon_1', ... });
 * ---
 * ```
 */

export { monitorClient, networkClient, resultClient, settingsClient, statusPageClient } from './client';

// Export all generated types for convenience
export * from '../gen/monitor/v1/monitor_pb';
export * from '../gen/network/v1/network_pb';
export * from '../gen/result/v1/result_pb';
export * from '../gen/common/v1/common_pb';
export * from '../gen/settings/v1/settings_pb';
export * from '../gen/statuspage/v1/statuspage_pb';

/**
 * Health check - verify API is reachable
 */
export async function healthCheck(): Promise<boolean> {
  try {
    const apiUrl = import.meta.env.PUBLIC_API_URL || 'http://localhost:8080';
    const response = await fetch(new URL('/health', apiUrl).toString(), {
      method: 'GET',
    });
    return response.ok;
  } catch (error) {
    console.warn('Health check failed:', error);
    return false;
  }
}
