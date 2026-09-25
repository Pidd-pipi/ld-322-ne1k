import { useCallback, useEffect, useState } from 'react';
import { Button, List, Modal, Select, Space, Switch, Tag, Typography, message } from 'antd';
import type { Device, Schedule } from '../../types/domain';
import { createSchedule, getSchedules, setScheduleEnabled } from '../../api/monitoring';

const { Text } = Typography;
const hours = Array.from({ length: 24 }, (_, i) => ({ value: i, label: `${String(i).padStart(2, '0')} 时` }));
const minutes = Array.from({ length: 60 }, (_, i) => ({ value: i, label: `${String(i).padStart(2, '0')} 分` }));
const serverMessage = (err: unknown) => (err as { response?: { data?: { message?: string } } })?.response?.data?.message;

export default function ScheduleModal({ device, onClose }: { device?: Device; onClose: () => void }) {
  const [schedules, setSchedules] = useState<Schedule[]>([]);
  const [hour, setHour] = useState(8);
  const [minute, setMinute] = useState(0);
  const [action, setAction] = useState<'on' | 'off'>('on');
  const [saving, setSaving] = useState(false);
  const load = useCallback(async () => {
    if (!device) return;
    try {
      setSchedules(await getSchedules(device.id));
    } catch {
      message.error('定时任务加载失败');
    }
  }, [device]);
  useEffect(() => {
    if (device) void load();
  }, [device, load]);
  const create = async () => {
    if (!device) return;
    setSaving(true);
    try {
      await createSchedule({ deviceId: device.id, hour, minute, action });
      message.success(`已创建每天 ${String(hour).padStart(2, '0')}:${String(minute).padStart(2, '0')} 的定时任务`);
      await load();
    } catch (err) {
      message.warning(serverMessage(err) ?? '定时任务创建失败，请先登录演示账号');
    } finally {
      setSaving(false);
    }
  };
  const toggle = async (row: Schedule, enabled: boolean) => {
    try {
      await setScheduleEnabled(row.id, enabled);
      message.success(enabled ? '任务已启用' : '任务已停用，同一时刻可重新安排');
      await load();
    } catch (err) {
      message.error(serverMessage(err) ?? '任务状态更新失败');
    }
  };
  return (
    <Modal title={`${device?.name ?? ''} · 定时任务`} open={!!device} onCancel={onClose} footer={null} destroyOnClose>
      <Space.Compact block>
        <Select aria-label="小时" value={hour} options={hours} onChange={setHour} style={{ width: 90 }} />
        <Select aria-label="分钟" value={minute} options={minutes} onChange={setMinute} style={{ width: 90 }} showSearch />
        <Select aria-label="动作" value={action} onChange={setAction} style={{ width: 90 }} options={[{ value: 'on', label: '开启' }, { value: 'off', label: '关闭' }]} />
        <Button type="primary" loading={saving} onClick={create}>新建任务</Button>
      </Space.Compact>
      <List
        style={{ marginTop: 16 }}
        dataSource={schedules}
        locale={{ emptyText: '暂无定时任务' }}
        renderItem={(row) => (
          <List.Item actions={[<Switch key="toggle" checked={row.enabled} onChange={(checked) => toggle(row, checked)} />]}>
            <List.Item.Meta
              title={<Space><Text strong>每天 {String(row.hour).padStart(2, '0')}:{String(row.minute).padStart(2, '0')}</Text><Tag color={row.action === 'on' ? 'green' : 'default'}>{row.action === 'on' ? '开启' : '关闭'}</Tag><Text type={row.enabled ? 'success' : 'secondary'}>{row.enabled ? '启用中' : '已停用'}</Text></Space>}
              description={row.lastRunAt ? `最近执行：${new Date(row.lastRunAt).toLocaleString()}` : '尚未执行'}
            />
          </List.Item>
        )}
      />
    </Modal>
  );
}
