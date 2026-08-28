import { useEffect, useState } from "react";
import { getChannels, getServers, createServer, createChannel, createInvite, joinServer, getMembers} from "./api";
import type { Channel, Server, User, Member } from "./types";
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

  const [selectedServerId, setSelectedServerId] =
    useState<number | null>(null);

  const [selectedChannelId, setSelectedChannelId] =
    useState<number | null>(null);

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
        console.error(
          "Could not load members:",
          error,
        );

        setMembers([]);
      }
    }

    loadMembers();
  }, [selectedServerId]);

  async function handleJoinServer(code: string) {

    const server = await joinServer(code);

    setServers((currentServers) => { 
      const alreadyExists = currentServers.some(
        (currentServer) => currentServer.id === server.id, 
      );

    if (alreadyExists) { 
      return currentServers; 
    } 

    return [ ...currentServers, server]; 
    }); 

    setSelectedServerId(server.id); 
  }

  async function handleCreateInvite(
    serverID: number,
  ) {
    return createInvite(serverID);
  }

  async function handleCreateServer(name: string) {
    const server = await createServer(name);

    setServers((currentServers) => [ ...currentServers, server, ]);

    setSelectedServerId(server.id); 
  }

  function handleSelectServer(serverID: number) {
    setSelectedServerId(serverID);
  }

  function handleSelectChannel(channelID: number) {
    setSelectedChannelId(channelID);
  }

  async function handleCreateChannel( name: string, type: "text" | "voice", ) {

    if (selectedServerId === null) {
      return; 
    }

    const channel = await createChannel( 
      selectedServerId, 
      name, 
      type,
    );

    setChannels((currentChannels) => [ 
      ...currentChannels, channel, 
    ]);

    setSelectedChannelId(channel.id); 
  }

  const selectedChannel =
    channels.find(
      (channel) => channel.id === selectedChannelId,
    ) ?? null;

  return (
    <div className="app">
      <ServerSidebar
        servers={servers}
        selectedServerId={selectedServerId}
        onSelectServer={handleSelectServer}
        onLogout={onLogout}
        onCreateServer={() => setShowCreateServer(true)}
        onJoinServer={() => setShowJoinServer(true)}
      />

      <div className="channel-area">
        {selectedServerId !== null && (
          <ChannelSidebar
            channels={channels}
            selectedChannelId={selectedChannelId}
            onSelectChannel={handleSelectChannel}
            onCreateChannel={() => setShowCreateChannel(true)}
            onInvite={() => setShowInvite(true)}
            serverId={selectedServerId}
          />
        )}

        <UserPanel
          user={user}
          onOpenProfile={() => setShowProfile(true)}
          onLogout={onLogout}
        />
      </div>


      <ChatWindow
        user={user}
        channel={selectedChannel}
      />

      <MembersSidebar members={members} />

      {showCreateServer && (
        <CreateServerModal
          onClose={() => setShowCreateServer(false)}
          onCreate={handleCreateServer}
        />
      )}

      {showCreateChannel && (
        <CreateChannelModal
          onClose={() => setShowCreateChannel(false)}
          onCreate={handleCreateChannel}
        />
      )}

      {showInvite && selectedServerId !== null && (
        <InviteModal
          serverID={selectedServerId}
          onClose={() => setShowInvite(false)}
          onCreateInvite={handleCreateInvite}
        />
      )}

      {showJoinServer && (
        <JoinServerModal
          onClose={() => setShowJoinServer(false)}
          onJoin={handleJoinServer}
        />
      )}

      {showProfile && (
        <ProfileModal
          user={user}
          onClose={() => setShowProfile(false)}
          onUserUpdate={onUserUpdate}
        />
      )}


    </div>
  );
}

export default Chat;
