import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { processService } from '../services/api';
import { useAppStore } from '../store/useAppStore';

// The queryFn ternary returned the service's real result on one branch and a
// hand-written literal on the other. TypeScript widened that union to `{}`,
// so every property access on the result failed once the processService
// facade stopped being typed `any`. Deriving the fallback from the service's
// own signature keeps both branches the same shape.
const NOTIFICATION_POLL_INTERVAL_MS = 60_000;

type NotificationsResult = Awaited<ReturnType<typeof processService.listNotifications>>;

export const useNotifications = () => {
  const user = useAppStore((state) => state.user);
  const queryClient = useQueryClient();

  const query = useQuery({
    queryKey: ['notifications', user?.id],
    queryFn: ({ signal }) =>
      user
        ? processService.listNotifications(user.username, signal)
        : Promise.resolve({ notifications: [], error: undefined } as NotificationsResult),
    enabled: !!user,
    // A fallback for when the SSE stream is not delivering; a minute is
    // often enough for that, and half the polling load of the 30s it was.
    refetchInterval: NOTIFICATION_POLL_INTERVAL_MS,
  });

  const markAsRead = useMutation({
    mutationFn: (id: string) => processService.markAsRead(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
    },
  });

  const markAllAsRead = useMutation({
    mutationFn: () => user ? processService.markAllAsRead(user.username) : Promise.resolve({}),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
    },
  });

  const deleteNotification = useMutation({
    mutationFn: (id: string) => processService.deleteNotification(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
    },
  });

  const unreadCount = (query.data?.notifications || []).filter((n) => !n.is_read).length;

  return {
    ...query,
    notifications: query.data?.notifications || [],
    unreadCount,
    markAsRead,
    markAllAsRead,
    deleteNotification,
  };
};
