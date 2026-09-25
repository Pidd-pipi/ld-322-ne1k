import client from './client'; import type { ApiResponse, DeviceSchedule } from '../types/domain';
export const listSchedules = async (greenhouseId: number) => (await client.get<ApiResponse<DeviceSchedule[]>>('/schedules', { params: { greenhouse_id: greenhouseId } })).data.data;
export const createSchedule = async (payload: { deviceId: number; hour: number; minute: number; action: 'on' | 'off' }) => (await client.post<ApiResponse<DeviceSchedule>>('/schedules', payload)).data.data;
export const setScheduleEnabled = async (id: number, enabled: boolean) => (await client.patch<ApiResponse<DeviceSchedule>>(`/schedules/${id}/status`, { enabled })).data.data;
