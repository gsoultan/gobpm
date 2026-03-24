import { AUTH_STORAGE_KEY } from "./config";

type StorageState = {
  state?: {
    token?: string;
  };
};

const parseStorageState = (): StorageState | null => {
  const storage = localStorage.getItem(AUTH_STORAGE_KEY);
  if (!storage) {
    return null;
  }

  try {
    return JSON.parse(storage) as StorageState;
  } catch (error) {
    console.error("Error parsing auth storage", error);
    return null;
  }
};

export const getAuthToken = (): string | null => {
  const state = parseStorageState();
  if (!state?.state?.token) {
    return null;
  }

  return state.state.token;
};

export const getAuthHeaders = (): Record<string, string> => {
  const token = getAuthToken();
  if (!token) {
    return {};
  }

  return { Authorization: `Bearer ${token}` };
};
