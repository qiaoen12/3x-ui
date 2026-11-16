import { Modal, Tag } from 'antd';
import { useTranslation } from 'react-i18next';
import { formatPanelVersion } from '@/lib/panel-version';
import ManagedUpdatesNotice from './ManagedUpdatesNotice';

export interface PanelUpdateInfo {
  channel?: string;
  currentVersion: string;
  latestVersion: string;
  currentCommit?: string;
  latestCommit?: string;
  updateAvailable: boolean;
}

interface BusyEvent {
  busy: boolean;
  tip?: string;
}

interface PanelUpdateModalProps {
  open: boolean;
  info: PanelUpdateInfo;
  devChannelEnable?: boolean;
  onChannelChange?: (dev: boolean) => void | Promise<void>;
  onClose: () => void;
  onBusy: (e: BusyEvent) => void;
}

export default function PanelUpdateModal({ open, info, onClose }: PanelUpdateModalProps) {
  const { t } = useTranslation();
  return (
    <Modal open={open} title={t('pages.index.updatePanel')} footer={null} onCancel={onClose}>
      <ManagedUpdatesNotice />
      <span>{t('pages.index.currentPanelVersion')} </span>
      <Tag color="green">
        {formatPanelVersion(window.X_UI_CUR_VER || info.currentVersion) || '?'}
      </Tag>
    </Modal>
  );
}
