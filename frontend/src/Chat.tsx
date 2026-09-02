import { useEffect, useRef, useState, useCallback } from "react";
import {
  getChannels,
  getServers,
  createServer,
  createChannel,
  createInvite,
  joinServer,
  getMembers,
  API_URL,
  getUnread,
  markChannelRead,
  type ChannelUnread,
} from "./api";

import type {
  Channel,
  Server,
  User,
  Member,
  VoiceParticipant,
  Message,
  RealtimeEvent,
  VoiceState,
} from "./types";

import ServerSidebar from "./ServerSidebar";
import ChannelSidebar from "./ChannelSidebar";
import ChatWindow from "./ChatWindow";
import CreateServerModal from "./CreateServerModal";
import CreateChannelModal from "./CreateChannelModal";
import InviteModal from "./InviteModal";
import JoinServerModal from "./JoinServerModal";
import MembersSidebar from "./MembersSidebar";
import UserPanel from "./UserPanel";
import ProfileModal from "./ProfileModal";
import VoiceConnection from "./VoiceConnection";


type ChatProps = {
  user: User;
  onLogout: () => void;
  onUserUpdate: (user: User) => void;
};




function Chat({ user, onLogout, onUserUpdate }: ChatProps) {
  const [servers, setServers] = useState<Server[]>([]);
  const [channels, setChannels] = useState<Channel[]>([]);
  const [showCreateServer, setShowCreateServer] = useState(false);
  const [showCreateChannel, setShowCreateChannel] = useState(false);
  const [showInvite, setShowInvite] = useState(false);
  const [showJoinServer, setShowJoinServer] = useState(false);
  const [members, setMembers] = useState<Member[]>([]);
  const [showProfile, setShowProfile] = useState(false);
  
  const [realtimeMessageEvent, setRealtimeMessageEvent] =
    useState<RealtimeEvent | null>(null);



  const [selectedServerId, setSelectedServerId] =
    useState<number | null>(null);

  const [selectedChannelId, setSelectedChannelId] =
    useState<number | null>(null);

  const [activeVoiceChannelId, setActiveVoiceChannelId] =
    useState<number | null>(null);

  const [muted, setMuted] = useState(false);
  const [deafened, setDeafened] = useState(false);

  const [voiceStates, setVoiceStates] = useState<VoiceState[]>([]);

  const [unreadChannels, setUnreadChannels] =
    useState<Record<number, ChannelUnread>>({});  

  /*
   * The realtime WebSocket for the currently selected server.
   */
  const realtimeSocketRef = useRef<WebSocket | null>(null);
  const userRealtimeSocketRef = useRef<WebSocket | null>(null);


  useEffect(() => {
    const WS_URL = API_URL.replace(/^http/, "ws");

    const socket = new WebSocket(
      `${WS_URL}/ws/realtime`
    );

    userRealtimeSocketRef.current = socket;

    socket.onopen = () => {
      console.log(
        "User realtime connected"
      );
    };

    socket.onmessage = (event) => {
      try {
        const message: RealtimeEvent =
          JSON.parse(event.data);

        console.log(
          "user realtime event:",
          message
        );

        if (message.type !== "unread_update") {
          return;
        }

        const update = message.data;

        if (update.user_id === user.id) {
          return;
        }

        setUnreadChannels((current) => ({
          ...current,
          [update.channel_id]: {
            channel_id: update.channel_id,
            server_id: update.server_id,
            unread_count:
              (current[update.channel_id]
                ?.unread_count ?? 0) + 1,
            last_message_id:
              update.message_id,
            last_read_message_id:
              current[update.channel_id]
                ?.last_read_message_id ?? 0,
          },
        }));
      } catch (error) {
        console.error(
          "Could not process user realtime event:",
          error,
        );
      }
    };

    socket.onerror = (error) => {
      console.error(
        "User realtime error:",
        error,
      );
    };

    socket.onclose = () => {
      console.log(
        "User realtime disconnected"
      );

      if (
        userRealtimeSocketRef.current === socket
      ) {
        userRealtimeSocketRef.current = null;
      }
    };

    return () => {
      socket.close();

      if (
        userRealtimeSocketRef.current === socket
      ) {
        userRealtimeSocketRef.current = null;
      }
    };
  }, [user.id]);

  useEffect(() => {
    async function loadUnread() {
      try {
        const response = await getUnread();

        const unread: Record<number, ChannelUnread> = {};

        for (const channel of response.channels) {
          unread[channel.channel_id] = channel;
        }

        setUnreadChannels(unread);
      } catch (error) {
        console.error(
          "Could not load unread messages:",
          error,
        );
      }
    }

    loadUnread();
  }, []);


  /*
   * -------------------------
   * Load servers
   * -------------------------
   */

  useEffect(() => {
    async function loadServers() {
      try {
        const servers = await getServers();

        setServers(servers);

        if (servers.length > 0) {
          setSelectedServerId(servers[0].id);
        }
      } catch (error) {
        console.error("Could not load servers:", error);
      }
    }

    loadServers();
  }, []);

  /*
   * -------------------------
   * Load channels
   * -------------------------
   */

  useEffect(() => {
    if (selectedServerId === null) {
      setChannels([]);
      setSelectedChannelId(null);
      return;
    }

    const serverId = selectedServerId;

    async function loadChannels() {
      try {
        const channels = await getChannels(serverId);

        setChannels(channels);

        if (channels.length > 0) {
          setSelectedChannelId(channels[0].id);
        } else {
          setSelectedChannelId(null);
        }
      } catch (error) {
        console.error("Could not load channels:", error);
        setChannels([]);
        setSelectedChannelId(null);
      }
    }

    loadChannels();
  }, [selectedServerId]);

  /*
   * -------------------------
   * Load members
   * -------------------------
   */

  useEffect(() => {
    if (selectedServerId === null) {
      setMembers([]);
      return;
    }

    const serverID = selectedServerId;

    async function loadMembers() {
      try {
        const members = await getMembers(serverID);

        setMembers(members);
      } catch (error) {
        console.error("Could not load members:", error);
        setMembers([]);
      }
    }

    loadMembers();
  }, [selectedServerId]);

  /*
   * -------------------------
   * Realtime server socket
   * -------------------------
   *
   * This socket is NOT the SFU socket.
   *
   * It is responsible for:
   * - messages
   * - voice presence
   * - future server events
   *
   * The SFU remains responsible for:
   * - WebRTC
   * - audio
   * - tracks
   */
  const refreshUnread = useCallback(async () => {
    try {
      const response = await getUnread();

      const unread: Record<number, ChannelUnread> = {};

      for (const channel of response.channels) {
        unread[channel.channel_id] = channel;
      }

      setUnreadChannels(unread);
    } catch (error) {
      console.error(
        "Could not refresh unread messages:",
        error,
      );
    }
  }, []);

  useEffect(() => {
    if (selectedServerId === null) {
      return;
    }

    const serverID = selectedServerId;

    /*
     * Close the previous realtime socket first.
     */
    if (realtimeSocketRef.current) {
      realtimeSocketRef.current.close();
      realtimeSocketRef.current = null;
    }


    const WS_URL = API_URL.replace(/^http/, "ws");

    const socket = new WebSocket(
      `${WS_URL}/ws/servers/${serverID}`
    );

    realtimeSocketRef.current = socket;

    socket.onopen = () => {
      console.log(
        `Realtime connected to server ${serverID}`,
      );
    };

    socket.onmessage = (event) => {
      console.log("realtime raw event:" , event.data)

      try {
        const message: RealtimeEvent =
          JSON.parse(event.data);

        switch (message.type) {
          case "voice_state": {
            setVoiceStates(
              Array.isArray(message.data)
                ? message.data
                : [],
            );

            break;
          }

          case "voice_join": {
            setVoiceStates((current) => {
              const withoutUser = current.filter(
                (state) =>
                  state.user_id !== message.data.user_id,
              );

              return [
                ...withoutUser,
                message.data,
              ];
            });

            break;
          }

          case "voice_leave": {
            setVoiceStates((current) =>
              current.filter(
                (state) =>
                  !(
                    state.user_id === message.data.user_id &&
                    state.channel_id ===
                      message.data.channel_id
                  ),
              ),
            );

            break;
          }
          case "voice_update": {
            setVoiceStates((current) =>
              current.map((state) =>
                state.user_id === message.data.user_id &&
                state.channel_id === message.data.channel_id
                  ? message.data
                  : state,
              ),
            );

            break;
          }
          case "user_updated": {
            const updatedUser = message.data;

            setMembers((currentMembers) =>
              currentMembers.map((member) =>
                member.id === updatedUser.id
                  ? {
                      ...member,
                      username: updatedUser.username,
                      avatar_url: updatedUser.avatar_url,
                    }
                  : member,
              ),
            );

            if (updatedUser.id === user.id) {
              onUserUpdate(updatedUser);
            }

            break;
          }

          case "message_created": {
            setRealtimeMessageEvent(message);

            const newMessage = message.data;

            if (newMessage.user_id === user.id) {
              break;
            }

            if (newMessage.channel_id === selectedChannelId) {
              break;
            }

            refreshUnread();

            break;
          }

          case "message_deleted": {
            setRealtimeMessageEvent(message);
            break;
          }
          case "message_updated":
          case "message_deleted": {
            setRealtimeMessageEvent(message);
            break;
          }
        }
      } catch (error) {
        console.error(
          "Could not process realtime event:",
          error,
        );
      }
    };

    socket.onerror = (error) => {
      console.error(
        "Realtime WebSocket error:",
        error,
      );
    };

    socket.onclose = () => {
      console.log(
        `Realtime disconnected from server ${serverID}`,
      );

      if (
        realtimeSocketRef.current === socket
      ) {
        realtimeSocketRef.current = null;
      }
    };

    return () => {
      /*
       * Closing this socket causes the Go server to
       * LeaveAllVoice() for this user.
       */
      socket.close();

      if (
        realtimeSocketRef.current === socket
      ) {
        realtimeSocketRef.current = null;
      }
    };
  }, [selectedServerId, selectedChannelId, refreshUnread]);

  const handleLatestMessage = useCallback(
    async (channelID: number, messageID: number) => {
      try {
        await markChannelRead(channelID, messageID);

        setUnreadChannels((current) => {
          const next = { ...current };

          delete next[channelID];

          return next;
        });
      } catch (error) {
        console.error(
          "Could not mark channel as read:",
          error,
        );
      }
    },
    [],
  );
  
  function handleUserUpdate(updatedUser: User) {
    // Update the user in App.tsx
    onUserUpdate(updatedUser);

    // Update this server's member list immediately
    setMembers((currentMembers) =>
      currentMembers.map((member) =>
        member.id === updatedUser.id
          ? {
              ...member,
              avatar_url: updatedUser.avatar_url,
            }
          : member,
      ),
    );
  }


  /*
   * -------------------------
   * Voice channel
   * -------------------------
   */

  function handleJoinVoiceChannel(channelId: number) {
    const socket = realtimeSocketRef.current;
    

    if (
      !socket ||
      socket.readyState !== WebSocket.OPEN
    ) {
      console.error(
        "Realtime socket is not connected",
      );
      return;
    }

    setActiveVoiceChannelId((current) => {
      if (current === channelId) {
        socket.send(
          JSON.stringify({
            type: "voice_leave",
            data: {
              channel_id: channelId,
            },
          }),
        );

        setMuted(false);
        setDeafened(false);

        return null;
      }

      if (current !== null) {
        socket.send(
          JSON.stringify({
            type: "voice_leave",
            data: {
              channel_id: current,
            },
          }),
        );

        setMuted(false);
        setDeafened(false);
      }

      socket.send(
        JSON.stringify({
          type: "voice_join",
          data: {
            channel_id: channelId,
          },
        }),
      );

      return channelId;
    });
  }

  function sendVoiceStateUpdate(
    nextMuted: boolean,
    nextDeafened: boolean,
  ) {
    const socket = realtimeSocketRef.current;

    if (
      !socket ||
      socket.readyState !== WebSocket.OPEN ||
      activeVoiceChannelId === null
    ) {
      return;
    }

    socket.send(
      JSON.stringify({
        type: "voice_update",
        data: {
          channel_id: activeVoiceChannelId,
          muted: nextMuted,
          deafened: nextDeafened,
        },
      }),
    );
  }
  

  /*
   * -------------------------
   * Server
   * -------------------------
   */

  function handleSelectServer(serverID: number) {
    if (serverID === selectedServerId) {
      return;
    }

    setActiveVoiceChannelId(null);
    setMuted(false);
    setDeafened(false);
    setVoiceStates([]);
    setSelectedServerId(serverID);
  }

  /*
   * -------------------------
   * Channel
   * -------------------------
   */

  function handleSelectChannel(
    channelID: number,
  ) {
    setSelectedChannelId(channelID);
  }

  /*
   * -------------------------
   * Join server
   * -------------------------
   */

  async function handleJoinServer(
    code: string,
  ) {
    const server = await joinServer(code);

    setServers((currentServers) => {
      const alreadyExists =
        currentServers.some(
          (currentServer) =>
            currentServer.id === server.id,
        );

      if (alreadyExists) {
        return currentServers;
      }

      return [
        ...currentServers,
        server,
      ];
    });

    setSelectedServerId(server.id);
  }

  /*
   * -------------------------
   * Invite
   * -------------------------
   */

  async function handleCreateInvite(
    serverID: number,
  ) {
    return createInvite(serverID);
  }

  /*
   * -------------------------
   * Create server
   * -------------------------
   */

  async function handleCreateServer(
    name: string,
  ) {
    const server = await createServer(name);

    setServers((currentServers) => [
      ...currentServers,
      server,
    ]);

    setSelectedServerId(server.id);
  }

  /*
   * -------------------------
   * Create channel
   * -------------------------
   */

  async function handleCreateChannel(
    name: string,
    type: "text" | "voice",
  ) {
    if (selectedServerId === null) {
      return;
    }

    const channel = await createChannel(
      selectedServerId,
      name,
      type,
    );

    setChannels((currentChannels) => [
      ...currentChannels,
      channel,
    ]);

    setSelectedChannelId(channel.id);
  }

  const selectedChannel =
    channels.find(
      (channel) =>
        channel.id === selectedChannelId,
    ) ?? null;

  const voiceParticipants: VoiceParticipant[] =
    (voiceStates ?? [])
      .map((state) => {
        const member = members.find(
          (member) =>
            member.id === state.user_id,
        );

        if (!member) {
          return null;
        }

        return {
          id: member.id,
          channelId: state.channel_id,
          username: member.username,

          // Use the avatar URL if your Member type has it.
          avatar_url:
            member.avatar_url ?? null,

          muted:
            state.user_id === user.id
              ? muted
              : state.muted,

          deafened:
            state.user_id === user.id
              ? deafened
              : state.deafened,


          speaking: false,
        };
      })
      .filter(
        (
          participant,
        ): participant is VoiceParticipant =>
          participant !== null,
      );

  function toggleMute() {
    if (activeVoiceChannelId === null) {
      return;
    }

    const nextMuted = !muted;

    setMuted(nextMuted);

    sendVoiceStateUpdate(
      nextMuted,
      deafened,
    );
  }

  function toggleDeafen() {
    if (activeVoiceChannelId === null) {
      return;
    }

    const nextDeafened = !deafened;

    /*
    * Deafening automatically mutes you.
    */
    const nextMuted =
      nextDeafened ? true : muted;

    setDeafened(nextDeafened);
    setMuted(nextMuted);

    sendVoiceStateUpdate(
      nextMuted,
      nextDeafened,
    );
  }

  const unreadServerIds = new Set(
    Object.values(unreadChannels).map(
      (channel) => channel.server_id,
    ),
  );

  return (
    <div className="app">
      <ServerSidebar
        servers={servers}
        selectedServerId={selectedServerId}
        onSelectServer={handleSelectServer}
        onLogout={onLogout}
        onCreateServer={() =>
          setShowCreateServer(true)
        }
        onJoinServer={() =>
          setShowJoinServer(true)
        }
        unreadServerIds={unreadServerIds}
      />

      <div className="channel-area">
        {selectedServerId !== null && (
          <ChannelSidebar
            channels={channels}
            selectedChannelId={
              selectedChannelId
            }
            onSelectChannel={
              handleSelectChannel
            }
            onCreateChannel={() =>
              setShowCreateChannel(true)
            }
            onInvite={() =>
              setShowInvite(true)
            }
            serverId={selectedServerId}
            activeVoiceChannelId={
              activeVoiceChannelId
            }
            onJoinVoiceChannel={
              handleJoinVoiceChannel
            }
            voiceParticipants={
              voiceParticipants
            }
            unreadChannels={unreadChannels}
          />
        )}

        <UserPanel
          user={user}
          muted={muted}
          deafened={deafened}
          onToggleMute={toggleMute}
          onToggleDeafen={toggleDeafen}
          onOpenProfile={() =>
            setShowProfile(true)
          }
          onLogout={onLogout}
        />
      </div>

      <ChatWindow
        user={user}
        channel={selectedChannel}
        members={members}
        realtimeEvent={realtimeMessageEvent}
        onLatestMessage={handleLatestMessage}
      />

      <MembersSidebar
        members={members}
      />

      {showCreateServer && (
        <CreateServerModal
          onClose={() =>
            setShowCreateServer(false)
          }
          onCreate={handleCreateServer}
        />
      )}

      {showCreateChannel && (
        <CreateChannelModal
          onClose={() =>
            setShowCreateChannel(false)
          }
          onCreate={handleCreateChannel}
        />
      )}

      {showInvite &&
        selectedServerId !== null && (
          <InviteModal
            serverID={selectedServerId}
            onClose={() =>
              setShowInvite(false)
            }
            onCreateInvite={
              handleCreateInvite
            }
          />
        )}

      {showJoinServer && (
        <JoinServerModal
          onClose={() =>
            setShowJoinServer(false)
          }
          onJoin={handleJoinServer}
        />
      )}

      {showProfile && (
        <ProfileModal
          user={user}
          onClose={() =>
            setShowProfile(false)
          }
          onUserUpdate={handleUserUpdate}
        />
      )}

      {activeVoiceChannelId !== null && (
        <VoiceConnection
          channelId={activeVoiceChannelId}
          muted={muted}
          deafened={deafened}
        />
      )}
    </div>
  );
}

export default Chat;
