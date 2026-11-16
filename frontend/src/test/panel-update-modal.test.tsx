import { fireEvent, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import PanelUpdateModal from '@/pages/index/PanelUpdateModal';
import { HttpUtil } from '@/utils';
import { renderWithProviders } from './test-utils';

describe('ops managed panel updates', () => {
  it('shows the installed version and operations notice, without an update or channel control', () => {
    const post = vi.spyOn(HttpUtil, 'post');
    const change = vi.fn();
    renderWithProviders(
      <PanelUpdateModal
        open
        info={{ currentVersion: '3.10.2', latestVersion: '99.0.0', updateAvailable: true }}
        devChannelEnable
        onChannelChange={change}
        onClose={vi.fn()}
        onBusy={vi.fn()}
      />,
    );
    expect(screen.getByText(/pinned operations workflow/)).toBeTruthy();
    expect(screen.getByText('v3.10.2')).toBeTruthy();
    expect(screen.queryByRole('switch')).toBeNull();
    for (const button of screen.getAllByRole('button')) fireEvent.click(button);
    expect(change).not.toHaveBeenCalled();
    expect(post).not.toHaveBeenCalled();
    post.mockRestore();
  });
});
