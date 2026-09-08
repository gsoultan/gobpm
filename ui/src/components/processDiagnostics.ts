/**
 * The shape of a per-element hint shown in the Smart Troubleshooter.
 *
 * This file used to also hold a whole second process validator — reachability,
 * gateway conditions, dead ends — that **nothing ever called**, and which named
 * steps by a `name` field the canvas does not set. Those checks now live in
 * `domain/processValidation.ts`, are wired into the Deploy button, and are
 * tested. Only the type the troubleshooter panel needs remains here.
 */
export interface DiagnosticResult {
  severity: 'error' | 'warning' | 'info';
  message: string;
  suggestion: string;
  quickFix?: () => void;
}
