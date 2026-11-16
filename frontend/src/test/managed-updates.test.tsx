import { fireEvent, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import VersionModal from '@/pages/index/VersionModal';
import NodeList from '@/pages/nodes/NodeList';
import { Status } from '@/models/status';
import { HttpUtil } from '@/utils';
import { renderWithProviders } from './test-utils';

const noop = () => {};

describe('ops managed Xray and node updates', () => {
  it('keeps the installed Xray version visible without downloads or update controls', () => {
    const get = vi.spyOn(HttpUtil, 'get');
    const post = vi.spyOn(HttpUtil, 'post');
    const status = new Status();
    status.xray.version = '26.10.10';
    renderWithProviders(<VersionModal open status={status} onClose={noop} onBusy={noop} />);
    expect(screen.getByText('26.10.10')).toBeTruthy();
    expect(screen.getByText(/pinned operations workflow/)).toBeTruthy();
    expect(screen.queryByRole('radio')).toBeNull();
    expect(get).not.toHaveBeenCalled();
    expect(post).not.toHaveBeenCalled();
    get.mockRestore();
    post.mockRestore();
  });
  it('disables per-node and bulk update actions even when an upstream version is available', () => {
    const update = vi.fn();
    renderWithProviders(
      <NodeList
        nodes={[
          { id: 1, name: 'test-node', enable: true, status: 'online', panelVersion: '3.10.2' },
        ]}
        latestVersion="99.0.0"
        isMobile={false}
        selectedIds={[1]}
        onSelectionChange={noop}
        onAdd={noop}
        onMtls={noop}
        onEdit={noop}
        onDelete={noop}
        onProbe={noop}
        onToggleEnable={noop}
        onUpdateNode={update}
        onUpdateSelected={update}
      />,
    );
    const buttons = screen
      .getAllByRole('button')
      .filter((b) => /update/i.test(b.textContent || b.getAttribute('aria-label') || ''));
    expect(buttons.length).toBeGreaterThan(0);
    for (const button of buttons) {
      expect((button as HTMLButtonElement).disabled).toBe(true);
      fireEvent.click(button);
    }
    expect(update).not.toHaveBeenCalled();
  });
});
