/**
 * useDesignerCollaboration — wraps the collaboration WebSocket channel for the
 * BPMN designer.
 *
 * FE-ARCH-5: extracted from useProcessDesigner so collaboration concerns are
 * isolated.  Handles:
 *   - Broadcasting the local cursor position to other participants
 *   - Applying remote node-move / node-update events to the local canvas
 */
import { useCallback, useEffect } from 'react';
import type { Node, Edge } from '@xyflow/react';
import type { ReactFlowInstance } from '@xyflow/react';
import { useCollaboration } from './useCollaboration';
import type { BPMNNodeData, BPMNEdgeData } from '../types/bpmn';

interface UseDesignerCollaborationParams {
  projectId: string | null | undefined;
  reactFlowInstance: ReactFlowInstance<Node<BPMNNodeData>, Edge<BPMNEdgeData>> | null;
  setNodes: React.Dispatch<React.SetStateAction<Node<BPMNNodeData>[]>>;
}

interface UseDesignerCollaborationReturn {
  remoteCursors: ReturnType<typeof useCollaboration>['remoteCursors'];
  broadcast: ReturnType<typeof useCollaboration>['broadcast'];
  /** Cursor-move handler — attach to the ReactFlow `onMouseMove` prop. */
  onMouseMove: (event: React.MouseEvent) => void;
}

export function useDesignerCollaboration({
  projectId,
  reactFlowInstance,
  setNodes,
}: UseDesignerCollaborationParams): UseDesignerCollaborationReturn {
  const { remoteCursors, remoteEvents, broadcast } = useCollaboration(projectId ?? undefined);

  // Apply incoming events from remote participants to the local canvas.
  useEffect(() => {
    remoteEvents.forEach((event) => {
      if (event.type === 'node_move') {
        const { id, position } = event.data as { id: string; position: { x: number; y: number } };
        setNodes((nds) => nds.map((n) => (n.id === id ? { ...n, position } : n)));
      } else if (event.type === 'node_update') {
        const { id, data: nodeData } = event.data as { id: string; data: Record<string, unknown> };
        setNodes((nds) =>
          nds.map((n) => (n.id === id ? { ...n, data: { ...n.data, ...nodeData } as BPMNNodeData } : n)),
        );
      }
    });
  }, [remoteEvents, setNodes]);

  const onMouseMove = useCallback(
    (event: React.MouseEvent) => {
      if (!reactFlowInstance) return;
      const { x, y } = reactFlowInstance.screenToFlowPosition({
        x: event.clientX,
        y: event.clientY,
      });
      broadcast('cursor', { x, y });
    },
    [reactFlowInstance, broadcast],
  );

  return { remoteCursors, broadcast, onMouseMove };
}

