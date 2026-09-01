import { useEffect, useRef, useState } from "react";
import {
  getChannels,
  getServers,
  createServer,
  createChannel,
  createInvite,
  joinServer,
  getMembers,
  API_URL,
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

  const [voiceStates, setVoiceStates] = useState<VoiceState[]>([]);

  /*
   * The realtime WebSocket for the currently selected server.
   */
  const realtimeSocketRef = useRef<WebSocket | null>(null);


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

          case "message_created":
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
  }, [selectedServerId]);


  

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

          muted: false,
          speaking: false,
        };
      })
      .filter(
        (
          participant,
        ): participant is VoiceParticipant =>
          participant !== null,
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
          />
        )}

        <UserPanel
          user={user}
          onOpenProfile={() =>
            setShowProfile(true)
          }
          onLogout={onLogout}
        />
      </div>

      <ChatWindow
        user={user}
        channel={selectedChannel}
        realtimeEvent={realtimeMessageEvent}
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
          onUserUpdate={onUserUpdate}
        />
      )}

      {activeVoiceChannelId !== null && (
        <VoiceConnection
          channelId={activeVoiceChannelId}
        />
      )}
    </div>
  );
}

export default Chat;
