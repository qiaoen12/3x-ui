import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import GeodataSection from '@/pages/index/GeodataSection';
import { HttpUtil } from '@/utils';

describe('ops managed Geo updates', () => {
  it('cannot fetch sources, enable a schedule, save, or restart a core', () => {
    const post = vi.spyOn(HttpUtil, 'post');
    render(<GeodataSection active onBusy={vi.fn()} onClose={vi.fn()} />);
    expect(screen.getByText(/pinned operations workflow/)).toBeTruthy();
    expect(screen.queryByRole('button')).toBeNull();
    expect(screen.queryByRole('textbox')).toBeNull();
    expect(post).not.toHaveBeenCalled();
    post.mockRestore();
  });
});
