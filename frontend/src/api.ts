import type { Channel, Message, Server, User, Member } from "./types";

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


export async function updateMessage(
  messageID: number,
  content: string,
): Promise<Message> {
  const response = await fetch(
    `${API_URL}/api/messages/${messageID}`,
    {
      method: "PATCH",
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
    throw new Error("Could not update message");
  }

  const data = await response.json();

  return data.message;
}

export async function deleteMessage(messageID: number): Promise<void> {
  const response = await fetch(
    `${API_URL}/api/messages/${messageID}`,
    {
      method: "DELETE",
      credentials: "include",
    },
  );

  if (!response.ok) {
    throw new Error("Could not delete message");
  }
}


export async function createServer(name: string): Promise<Server> {
  const response = await fetch(`${API_URL}/api/servers`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    credentials: "include",
    body: JSON.stringify({
      name,
    }),
  });

  if (!response.ok) {
    throw new Error("Could not create server");
  }

  const data = await response.json();

  return data.server;
}


export async function createChannel(
  serverID: number,
  name: string,
  type: "text" | "voice",
): Promise<Channel> {
  const response = await fetch(
    `${API_URL}/api/servers/${serverID}/channels`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
      body: JSON.stringify({
        name,
        type,
      }),
    },
  );

  if (!response.ok) {
    throw new Error("Could not create channel");
  }

  const data = await response.json();

  return data.channel;
}


export async function logout(): Promise<void> {
  const response = await fetch(`${API_URL}/api/logout`, {
    method: "POST",
    credentials: "include",
  });

  if (!response.ok) {
    throw new Error("Could not log out");
  }
}


export async function createInvite(
  serverID: number,
): Promise<string> {
  const response = await fetch(
    `${API_URL}/api/servers/${serverID}/invites`,
    {
      method: "POST",
      credentials: "include",
    },
  );

  if (!response.ok) {
    throw new Error("Could not create invite");
  }

  const data = await response.json();

  return data.code;
}

export async function joinServer(
  code: string,
): Promise<Server> {
  const response = await fetch(
    `${API_URL}/api/invites/${encodeURIComponent(code)}/join`,
    {
      method: "POST",
      credentials: "include",
    },
  );

  if (!response.ok) {
    throw new Error("Could not join server");
  }

  const data = await response.json();

  return data.server;
}


export async function getMembers(
  serverID: number,
): Promise<Member[]> {
  const response = await fetch(
    `${API_URL}/api/servers/${serverID}/members`,
    {
      credentials: "include",
    },
  );

  if (!response.ok) {
    throw new Error("Could not load members");
  }

  const data = await response.json();

  return data.members;
}


export async function register(
  username: string,
  password: string,
): Promise<User> {
  const response = await fetch(`${API_URL}/api/register`, {
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
    const message = await response.text();
    throw new Error(message || "Could not create account");
  }

  const data = await response.json();

  return data.user;
}
