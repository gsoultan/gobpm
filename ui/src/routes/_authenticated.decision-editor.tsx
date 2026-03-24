import { createFileRoute } from '@tanstack/react-router';
import { DecisionEditor } from '../pages/DecisionEditor';
import { z } from 'zod';

export const Route = createFileRoute('/_authenticated/decision-editor')({
  validateSearch: z.object({
    id: z.string().optional(),
    name: z.string().optional(),
    key: z.string().optional(),
  }),
  component: () => {
    const { id } = Route.useSearch();
    return <DecisionEditor definitionId={id} />;
  }
});
