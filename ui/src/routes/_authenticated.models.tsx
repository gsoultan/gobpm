import { createFileRoute } from '@tanstack/react-router';
import { Models } from '../pages/Models';
import { useAppStore } from '../store/useAppStore';
import { useEffect } from 'react';
import { z } from 'zod';

const modelsSearchSchema = z.object({
  tab: z.enum(['processes', 'decisions']).catch('processes'),
});

export const Route = createFileRoute('/_authenticated/models')({
  component: ModelsRoute,
  validateSearch: (search) => modelsSearchSchema.parse(search),
});

function ModelsRoute() {
  const { setActiveTab } = useAppStore();

  useEffect(() => {
    setActiveTab('models');
  }, [setActiveTab]);

  return <Models />;
}
