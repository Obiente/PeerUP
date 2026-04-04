/**
 * Status Pages Service
 * Manages public status pages linked to monitors through the Connect API.
 */

import { statusPageClient } from '../api/client';
import {
  CreateStatusPageRequest,
  DeleteStatusPageRequest,
  GetStatusPageRequest,
  RecordVisitRequest,
  UpdateStatusPageRequest,
} from '../../gen/statuspage/v1/statuspage_pb';

export interface StatusPage {
  id: string;
  name: string;
  slug: string;
  description?: string;
  monitorIds: string[];
  customDomain?: string;
  isPublished: boolean;
  displayIncidents: boolean;
  displayUptime: boolean;
  refreshInterval: number;
  createdAt: Date;
  updatedAt: Date;
  customCSS?: string;
  companyName?: string;
  companyLogo?: string;
}

export interface StatusPageSettings {
  theme: 'light' | 'dark' | 'auto';
  language: string;
  timezone: string;
  showResponseTime: boolean;
  showCheckFrequency: boolean;
}

export interface PublicStatusPageData {
  page: StatusPage;
  settings: StatusPageSettings;
  monitors: Array<{
    id: string;
    name: string;
    status: 'up' | 'down' | 'degraded' | 'unknown';
    lastUpdated: Date;
    uptime24h: number;
    uptime30d: number;
    averageResponseTime: number;
  }>;
  incidents: Array<{
    id: string;
    title: string;
    description: string;
    startTime: Date;
    endTime?: Date;
    status: 'investigating' | 'identified' | 'monitoring' | 'resolved';
    affectedServices: string[];
  }>;
}

function toLocalStatusPage(page: {
  id: string;
  title: string;
  slug: string;
  description: string;
  monitorIds: string[];
  customDomain: string;
  isActive: boolean;
  createdAt: bigint | number | string;
  updatedAt: bigint | number | string;
  logoUrl: string;
}): StatusPage {
  return {
    id: page.id,
    name: page.title,
    slug: page.slug,
    description: page.description,
    monitorIds: page.monitorIds,
    customDomain: page.customDomain || undefined,
    isPublished: page.isActive,
    displayIncidents: true,
    displayUptime: true,
    refreshInterval: 30,
    createdAt: new Date(Number(page.createdAt) * 1000),
    updatedAt: new Date(Number(page.updatedAt) * 1000),
    companyLogo: page.logoUrl || undefined,
  };
}

export async function createStatusPage(data: {
  name: string;
  slug: string;
  description?: string;
  monitorIds: string[];
  companyName?: string;
  customDomain?: string;
}): Promise<{ success: boolean; page?: StatusPage; error?: string }> {
  try {
    const created = await statusPageClient.createStatusPage(new CreateStatusPageRequest({
      title: data.name,
      slug: data.slug,
      description: data.description ?? '',
      monitorIds: data.monitorIds,
    }));

    return {
      success: true,
      page: {
        ...toLocalStatusPage(created),
        companyName: data.companyName,
        customDomain: created.customDomain || data.customDomain,
      },
    };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Failed to create status page',
    };
  }
}

export async function getStatusPage(slug: string): Promise<{
  success: boolean;
  page?: PublicStatusPageData;
  error?: string;
}> {
  try {
    const page = await statusPageClient.getStatusPage(new GetStatusPageRequest({
      identifier: { case: 'slug', value: slug },
    }));

    return {
      success: true,
      page: {
        page: toLocalStatusPage(page),
        settings: {
          theme: 'light',
          language: 'en',
          timezone: 'UTC',
          showResponseTime: true,
          showCheckFrequency: true,
        },
        monitors: page.monitorIds.map((monitorId) => ({
          id: monitorId,
          name: monitorId,
          status: page.uptime >= 99 ? 'up' as const : page.uptime >= 95 ? 'degraded' as const : 'down' as const,
          lastUpdated: new Date(Number(page.updatedAt) * 1000),
          uptime24h: page.uptime,
          uptime30d: page.uptime,
          averageResponseTime: 0,
        })),
        incidents: [],
      },
    };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Failed to fetch status page',
    };
  }
}

export async function updateStatusPage(
  id: string,
  data: Partial<StatusPage>
): Promise<{ success: boolean; page?: StatusPage; error?: string }> {
  try {
    const updated = await statusPageClient.updateStatusPage(new UpdateStatusPageRequest({
      id,
      title: data.name,
      slug: data.slug,
      description: data.description,
      monitorIds: data.monitorIds ?? [],
      isActive: data.isPublished,
      logoUrl: data.companyLogo,
    }));

    return {
      success: true,
      page: {
        ...toLocalStatusPage(updated),
        companyName: data.companyName,
      },
    };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Failed to update status page',
    };
  }
}

export async function publishStatusPage(id: string): Promise<{
  success: boolean;
  error?: string;
}> {
  try {
    await statusPageClient.updateStatusPage(new UpdateStatusPageRequest({
      id,
      isActive: true,
    }));
    return { success: true };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Failed to publish status page',
    };
  }
}

export async function deleteStatusPage(id: string): Promise<{
  success: boolean;
  error?: string;
}> {
  try {
    await statusPageClient.deleteStatusPage(new DeleteStatusPageRequest({ id }));
    return { success: true };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Failed to delete status page',
    };
  }
}

export async function getMonitorUptime(
  _monitorId: string,
  _days: number = 30
): Promise<{
  success: boolean;
  uptime?: number;
  incidents?: number;
  avgResponseTime?: number;
  error?: string;
}> {
  return {
    success: true,
    uptime: 100,
    incidents: 0,
    avgResponseTime: 0,
  };
}

export async function validateSlug(slug: string): Promise<{
  valid: boolean;
  message?: string;
}> {
  const slugRegex = /^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$/;
  if (!slugRegex.test(slug)) {
    return {
      valid: false,
      message: 'Slug must contain only lowercase letters, numbers, and hyphens',
    };
  }

  if (slug.length < 3) {
    return {
      valid: false,
      message: 'Slug must be at least 3 characters long',
    };
  }

  try {
    await statusPageClient.getStatusPage(new GetStatusPageRequest({
      identifier: { case: 'slug', value: slug },
    }));
    return {
      valid: false,
      message: 'Slug is already in use',
    };
  } catch {
    return { valid: true };
  }
}

export async function recordStatusPageVisit(id: string): Promise<void> {
  await statusPageClient.recordVisit(new RecordVisitRequest({ statusPageId: id }));
}
