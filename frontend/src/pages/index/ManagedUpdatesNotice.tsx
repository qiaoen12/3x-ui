import { Alert } from 'antd';
import { useTranslation } from 'react-i18next';

export default function ManagedUpdatesNotice() {
  const { t } = useTranslation();
  return (
    <Alert type="info" showIcon className="mb-12" title={t('pages.index.managedUpdatesNotice')} />
  );
}
