/**
 * useDesignerHistory — manages undo / redo history for the BPMN canvas.
 *
 * FE-ARCH-5: extracted from useProcessDesigner so that history state is
 * isolated in its own focused hook.
 *
 * The hook owns the history stack and exposes:
 *   - pushToHistory   — call after any structural canvas change
 *   - undo / redo     — navigate the stack
 *   - history / historyIndex — for toolbar button enable states
 */
import { useCallback, useRef, useState } from 'react';
import type { Edge, Node } from '@xyflow/react';
import type { BPMNNodeData, BPMNEdgeData } from '../types/bpmn';

type HistoryItem = { nodes: Node<BPMNNodeData>[]; edges: Edge<BPMNEdgeData>[] };

const MAX_HISTORY_STEPS = 50;

interface UseDesignerHistoryReturn {
  history: HistoryItem[];
  historyIndex: number;
  /** Register a structural change.  Ignored during undo/redo replays. */
  pushToHistory: (nodes: Node<BPMNNodeData>[], edges: Edge<BPMNEdgeData>[]) => void;
  undo: (setNodes: (n: Node<BPMNNodeData>[]) => void, setEdges: (e: Edge<BPMNEdgeData>[]) => void) => void;
  redo: (setNodes: (n: Node<BPMNNodeData>[]) => void, setEdges: (e: Edge<BPMNEdgeData>[]) => void) => void;
}

export function useDesignerHistory(): UseDesignerHistoryReturn {
  const [history, setHistory] = useState<HistoryItem[]>([]);
  const [historyIndex, setHistoryIndex] = useState(-1);
  /** Prevents re-registering a history entry during undo/redo replay. */
  const isUndoRedoRef = useRef(false);

  const pushToHistory = useCallback((nodes: Node<BPMNNodeData>[], edges: Edge<BPMNEdgeData>[]) => {
    if (isUndoRedoRef.current) return;

    setHistory((prev) => {
      const base = prev.slice(0, historyIndex + 1);
      const next = [...base, { nodes, edges }];
      // Keep the stack bounded to avoid unbounded memory growth.
      return next.length > MAX_HISTORY_STEPS ? next.slice(1) : next;
    });
    setHistoryIndex((i) => Math.min(i + 1, MAX_HISTORY_STEPS - 1));
  }, [historyIndex]);

  const undo = useCallback(
    (setNodes: (n: Node<BPMNNodeData>[]) => void, setEdges: (e: Edge<BPMNEdgeData>[]) => void) => {
      if (historyIndex <= 0) return;
      isUndoRedoRef.current = true;
      const prev = history[historyIndex - 1];
      setNodes(prev.nodes);
      setEdges(prev.edges);
      setHistoryIndex((i) => i - 1);
      setTimeout(() => { isUndoRedoRef.current = false; }, 0);
    },
    [history, historyIndex],
  );

  const redo = useCallback(
    (setNodes: (n: Node<BPMNNodeData>[]) => void, setEdges: (e: Edge<BPMNEdgeData>[]) => void) => {
      if (historyIndex >= history.length - 1) return;
      isUndoRedoRef.current = true;
      const next = history[historyIndex + 1];
      setNodes(next.nodes);
      setEdges(next.edges);
      setHistoryIndex((i) => i + 1);
      setTimeout(() => { isUndoRedoRef.current = false; }, 0);
    },
    [history, historyIndex],
  );

  return { history, historyIndex, pushToHistory, undo, redo };
}

