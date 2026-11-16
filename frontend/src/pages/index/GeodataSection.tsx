import ManagedUpdatesNotice from './ManagedUpdatesNotice';

interface GeodataSectionProps {
  active: boolean;
  onBusy: (e: { busy: boolean; tip?: string }) => void;
  onClose: () => void;
}

export default function GeodataSection(_props: GeodataSectionProps) {
  return <ManagedUpdatesNotice />;
}
