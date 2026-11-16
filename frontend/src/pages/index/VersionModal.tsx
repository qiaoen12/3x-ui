import { Modal, Tag } from 'antd';
import { useTranslation } from 'react-i18next';
import type { Status } from '@/models/status';
import ManagedUpdatesNotice from './ManagedUpdatesNotice';

interface BusyEvent {
  busy: boolean;
  tip?: string;
}

interface VersionModalProps {
  open: boolean;
  status: Status;
  onClose: () => void;
  onBusy: (e: BusyEvent) => void;
}

export default function VersionModal({ open, status, onClose }: VersionModalProps) {
  const { t } = useTranslation();
  return (
    <Modal open={open} title={t('pages.index.xrayUpdates')} footer={null} onCancel={onClose}>
      <ManagedUpdatesNotice />
      <span>Xray </span>
      <Tag>{status.xray.version}</Tag>
    </Modal>
  );
}
