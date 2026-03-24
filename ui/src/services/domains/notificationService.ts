import { requestJSON } from "../shared/rest";

export type Notification = {
  id: string;
  user_id: string;
  type: string;
  title: string;
  message: string;
  is_read: boolean;
  link?: string;
  created_at: string;
  project_id?: string;
  instance_id?: string;
};

type ListNotificationsResponse = {
  notifications?: Notification[];
  error?: string;
};

type GenericResponse = {
  error?: string;
};

export const notificationService = {
  async listNotifications(userId: string, signal?: AbortSignal) {
    const data = await requestJSON<ListNotificationsResponse>(`/notifications?user_id=${userId}`, { signal });
    return { notifications: data.notifications ?? [], error: data.error };
  },

  async markAsRead(id: string, signal?: AbortSignal) {
    return await requestJSON<GenericResponse>(`/notifications/${id}/read`, {
      method: "POST",
      signal,
    });
  },

  async markAllAsRead(userId: string, signal?: AbortSignal) {
    return await requestJSON<GenericResponse>(`/notifications/read-all?user_id=${userId}`, {
      method: "POST",
      signal,
    });
  },

  async deleteNotification(id: string, signal?: AbortSignal) {
    return await requestJSON<GenericResponse>(`/notifications/${id}`, {
      method: "DELETE",
      signal,
    });
  },
};
