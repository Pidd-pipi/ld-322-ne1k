import { Button, Card, InputNumber, List, Select, Space, Switch, Tag, message } from 'antd';
import { useState } from 'react';
import axios from 'axios';
import type { Device, DeviceSchedule } from '../../types/domain';
import { createSchedule, setScheduleEnabled } from '../../api/schedules';

const pad = (value: number) => String(value).padStart(2, '0');

export default function SchedulePanel({ devices, schedules, onUpdated }: { devices: Device[]; schedules: DeviceSchedule[]; onUpdated: () => void }) {
  const [deviceId, setDeviceId] = useState<number>();
  const [hour, setHour] = useState(8);
  const [minute, setMinute] = useState(0);
  const [action, setAction] = useState<'on' | 'off'>('on');
  const [saving, setSaving] = useState(false);
  const [pendingId, setPendingId] = useState<number>();

  const deviceName = (item: DeviceSchedule) => item.device?.name ?? devices.find(d => d.id === item.deviceId)?.name ?? `设备 #${item.deviceId}`;

  const create = async () => {
    if (!deviceId) { message.warning('请选择要定时的设备'); return; }
    try {
      setSaving(true);
      await createSchedule({ deviceId, hour, minute, action });
      message.success(`已创建每天 ${pad(hour)}:${pad(minute)} 的定时任务`);
      onUpdated();
    } catch (err) {
      if (axios.isAxiosError(err) && err.response?.data?.message) message.warning(err.response.data.message);
      else message.error('定时任务创建失败，请先登录演示账号');
    } finally { setSaving(false); }
  };

  const toggleEnabled = async (item: DeviceSchedule, enabled: boolean) => {
    try {
      setPendingId(item.id);
      await setScheduleEnabled(item.id, enabled);
      message.success(enabled ? '任务已启用' : '任务已停用，该时刻可重新安排');
      onUpdated();
    } catch { message.error('操作失败，请先登录演示账号'); } finally { setPendingId(undefined); }
  };

  return <Card title="定时任务" extra={<Button size="small" onClick={onUpdated}>刷新</Button>}>
    <Space wrap style={{ marginBottom: 12 }}>
      <Select aria-label="选择设备" placeholder="选择设备" style={{ width: 140 }} value={deviceId} options={devices.map(d => ({ value: d.id, label: d.name }))} onChange={setDeviceId} />
      <InputNumber aria-label="小时" min={0} max={23} value={hour} onChange={v => setHour(v ?? 0)} addonAfter="时" style={{ width: 92 }} />
      <InputNumber aria-label="分钟" min={0} max={59} value={minute} onChange={v => setMinute(v ?? 0)} addonAfter="分" style={{ width: 92 }} />
      <Select aria-label="执行动作" style={{ width: 90 }} value={action} options={[{ value: 'on', label: '开启' }, { value: 'off', label: '关闭' }]} onChange={setAction} />
      <Button type="primary" loading={saving} onClick={create}>新建任务</Button>
    </Space>
    <List size="small" locale={{ emptyText: '暂无定时任务' }} dataSource={schedules} renderItem={item => <List.Item actions={[<Switch key="enabled" checked={item.enabled} loading={pendingId === item.id} onChange={v => toggleEnabled(item, v)} />]}>
      <List.Item.Meta
        title={<span>{deviceName(item)} · 每天 {pad(item.hour)}:{pad(item.minute)} · {item.action === 'on' ? '开启' : '关闭'} <Tag color={item.enabled ? 'success' : 'default'}>{item.enabled ? '启用中' : '已停用'}</Tag></span>}
        description={item.lastRunAt ? `上次执行：${new Date(item.lastRunAt).toLocaleString('zh-CN')}` : '尚未执行'}
      />
    </List.Item>} />
  </Card>;
}
