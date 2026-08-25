import type { Channel, Message, Server, User } from "./types";

const API_URL = "http://localhost:8080";

type LoginResponse = {
  message: string;
  user: User;
};

export async function login(username: string, password: string): Promise<LoginResponse> {
  const response = await fetch(`${API_URL}/api/login`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    credentials: "include",
    body: JSON.stringify({
      username,
      password,
    }),
  });

  if (!response.ok) {
    throw new Error("Invalid username or password");
  }

  return response.json();
}

export async function getMe(): Promise<User> {
  const response = await fetch(`${API_URL}/api/me`, {
    credentials: "include",
  });

  if (!response.ok) {
    throw new Error("Not authenticated");
  }

  return response.json();
}




export async function getServers(): Promise<Server[]> {
  const response = await fetch(`${API_URL}/api/servers`, {
    credentials: "include",
  });

  if (!response.ok) {
    throw new Error("Could not load servers");
  }

  const data = await response.json();

  return data.servers;
}

export async function getChannels(serverID: number): Promise<Channel[]> {
  const response = await fetch(
    `${API_URL}/api/servers/${serverID}/channels`,
    {
      credentials: "include",
    },
  );

  if (!response.ok) {
    throw new Error("Could not load channels");
  }

  const data = await response.json();

  return data.channels;
}



export async function getMessages(
  channelID: number,
): Promise<Message[]> {
  const response = await fetch(
    `${API_URL}/api/channels/${channelID}/messages`,
    {
      credentials: "include",
    },
  );

  if (!response.ok) {
    throw new Error("Could not load messages");
  }

  const data = await response.json();

  return data.messages;
}

export async function sendMessage(
  channelID: number,
  content: string,
): Promise<Message> {
  const response = await fetch(
    `${API_URL}/api/channels/${channelID}/messages`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
      body: JSON.stringify({
        content,
      }),
    },
  );

  if (!response.ok) {
    throw new Error("Could not send message");
  }

  const data = await response.json();

  return data.message;
}

