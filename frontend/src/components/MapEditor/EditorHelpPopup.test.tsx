import { describe, it, expect } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { EditorHelpPopup } from './EditorHelpPopup';

describe('EditorHelpPopup', () => {
  it('ne rend rien quand open est false', () => {
    const { container } = render(<EditorHelpPopup open={false} onClose={() => {}} />);
    expect(container.innerHTML).toBe('');
  });

  it('rend le titre et le tableau quand open est true', () => {
    render(<EditorHelpPopup open={true} onClose={() => {}} />);
    expect(screen.getByText(/Raccourcis clavier/i)).toBeTruthy();
    expect(screen.getByRole('columnheader', { name: /^Commande$/i })).toBeTruthy();
    expect(screen.getByRole('columnheader', { name: /^Raccourci$/i })).toBeTruthy();
    expect(screen.getByRole('columnheader', { name: /^Contexte$/i })).toBeTruthy();
  });

  it('affiche toutes les entrées de raccourcis', () => {
    render(<EditorHelpPopup open={true} onClose={() => {}} />);
    expect(screen.getByText(/Sauvegarder la zone/i)).toBeTruthy();
    expect(screen.getByText(/Ctrl\+S/i)).toBeTruthy();
    expect(screen.getByText(/Annuler \(undo\)/i)).toBeTruthy();
    expect(screen.getByText(/Delete \/ Suppr/i)).toBeTruthy();
    expect(screen.getByText(/Clic gauche/i)).toBeTruthy();
    expect(screen.getAllByText(/Clic droit/i).length).toBe(2);
    expect(screen.getByText(/Double-clic/i)).toBeTruthy();
  });

  it('appelle onClose quand le bouton Fermer est cliqué', () => {
    let closed = false;
    render(<EditorHelpPopup open={true} onClose={() => { closed = true; }} />);
    const closeBtn = screen.getByRole('button', { name: /fermer/i });
    fireEvent.click(closeBtn);
    expect(closed).toBe(true);
  });

  it('appelle onClose quand Escape est pressé', () => {
    let closed = false;
    render(<EditorHelpPopup open={true} onClose={() => { closed = true; }} />);
    fireEvent.keyDown(window, { key: 'Escape' });
    expect(closed).toBe(true);
  });

  it('appelle onClose quand on clique sur l\'overlay', () => {
    let closed = false;
    const { container } = render(<EditorHelpPopup open={true} onClose={() => { closed = true; }} />);
    const overlay = container.querySelector('[class*="overlay"]') || container.firstChild;
    if (overlay) {
      fireEvent.click(overlay);
      expect(closed).toBe(true);
    }
  });
});
