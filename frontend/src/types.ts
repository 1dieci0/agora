export type User = {
  id: number;
  username: string;
};

export type Server = {
  id: number;
  name: string;
  owner_id: number;
};

export type Channel = {
  id: number;
  server_id: number;
  name: string;
  type: string;
};

export type Message = {
  id: number;
  channel_id: number;
  user_id: number;
  username: string;
  content: string;
  created_at: string;
};
