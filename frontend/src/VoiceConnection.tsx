import { useEffect, useRef } from "react";

type VoiceConnectionProps = {
  channelId: number;
  currentUserId: number;
};

type SignalMessage = {
  type:
    | "user_joined"
    | "user_left"
    | "offer"
    | "answer"
    | "ice_candidate";

  user_id?: number | string;
  from?: number | string;
  to?: number | string;
  sdp?: string;
  candidate?: RTCIceCandidateInit | string | null;
};

type Peer = {
  pc: RTCPeerConnection;
  audio: HTMLAudioElement;
  remoteStream: MediaStream;
};

/*
 * ============================================================
 * LOCAL TEST CONFIG
 * ============================================================
 *
 * Two browsers on the SAME PC do not need STUN or TURN.
 *
 * We deliberately use no iceServers here.
 *
 * We also use non-trickle ICE below:
 * wait until ICE gathering completes, then send the SDP.
 */
const RTC_CONFIG: RTCConfiguration = {
  iceServers: [],
  iceTransportPolicy: "all",
  bundlePolicy: "max-bundle",
  rtcpMuxPolicy: "require",
};

export default function VoiceConnection({
  channelId,
  currentUserId,
}: VoiceConnectionProps) {
  const wsRef = useRef<WebSocket | null>(null);

  const localStreamRef =
    useRef<MediaStream | null>(null);

  const peersRef =
    useRef<Map<number, Peer>>(new Map());

  const generationRef =
    useRef(0);

  const outgoingQueueRef =
    useRef<string[]>([]);

  const cancelledRef =
    useRef(false);

  useEffect(() => {
    const generation =
      ++generationRef.current;

    let cancelled = false;

    cancelledRef.current = false;

    console.log(
      "================================================"
    );

    console.log("VOICE START", {
      channelId,
      currentUserId,
      generation,
    });

    console.log(
      "================================================"
    );

    /*
     * ------------------------------------------------------------
     * USER ID
     * ------------------------------------------------------------
     */

    function toUserId(
      value: unknown
    ): number | null {
      const id =
        typeof value === "number"
          ? value
          : typeof value === "string"
            ? Number(value)
            : NaN;

      if (
        !Number.isFinite(id) ||
        id <= 0
      ) {
        return null;
      }

      return id;
    }

    function getRemoteUserId(
      message: SignalMessage
    ): number | null {
      if (
        message.type ===
          "user_joined" ||
        message.type ===
          "user_left"
      ) {
        return toUserId(
          message.user_id
        );
      }

      return toUserId(
        message.from
      );
    }

    /*
     * ------------------------------------------------------------
     * SIGNAL SEND
     * ------------------------------------------------------------
     */

    function sendSignal(
      message: Record<string, unknown>
    ) {
      const ws =
        wsRef.current;

      const serialized =
        JSON.stringify(message);

      if (!ws) {
        console.warn(
          "VOICE: websocket doesn't exist; queueing",
          message
        );

        outgoingQueueRef.current.push(
          serialized
        );

        return;
      }

      if (
        ws.readyState ===
        WebSocket.OPEN
      ) {
        console.log(
          "VOICE SIGNAL SEND:",
          message
        );

        ws.send(serialized);

        return;
      }

      if (
        ws.readyState ===
        WebSocket.CONNECTING
      ) {
        console.log(
          "VOICE: websocket connecting; queueing",
          message
        );

        outgoingQueueRef.current.push(
          serialized
        );

        return;
      }

      console.warn(
        "VOICE: websocket unavailable",
        {
          readyState:
            ws.readyState,
          message,
        }
      );
    }

    function flushOutgoingQueue() {
      const ws =
        wsRef.current;

      if (
        !ws ||
        ws.readyState !==
          WebSocket.OPEN
      ) {
        return;
      }

      while (
        outgoingQueueRef.current
          .length > 0
      ) {
        const serialized =
          outgoingQueueRef.current.shift();

        if (!serialized) {
          continue;
        }

        const message =
          JSON.parse(serialized);

        console.log(
          "VOICE: flushing queued signal",
          message
        );

        ws.send(serialized);
      }
    }

    /*
     * ------------------------------------------------------------
     * WAIT FOR ICE GATHERING
     * ------------------------------------------------------------
     *
     * For local debugging we deliberately wait for ICE gathering
     * to finish before sending the offer/answer.
     *
     * This means the SDP should contain:
     *
     *   a=candidate:...
     *
     * before it reaches the other browser.
     */

    function waitForIceGathering(
      pc: RTCPeerConnection,
      remoteUserId: number
    ): Promise<void> {
      if (
        pc.iceGatheringState ===
        "complete"
      ) {
        console.log(
          "VOICE: ICE already complete",
          remoteUserId
        );

        return Promise.resolve();
      }

      return new Promise(
        (resolve) => {
          const timeout =
            window.setTimeout(() => {
              console.warn(
                "VOICE: ICE gathering timeout",
                {
                  remoteUserId,
                  state:
                    pc.iceGatheringState,
                }
              );

              resolve();
            }, 5000);

          const check = () => {
            console.log(
              "VOICE: ICE GATHERING STATE",
              {
                remoteUserId,
                state:
                  pc.iceGatheringState,
              }
            );

            if (
              pc.iceGatheringState ===
              "complete"
            ) {
              window.clearTimeout(
                timeout
              );

              pc.removeEventListener(
                "icegatheringstatechange",
                check
              );

              resolve();
            }
          };

          pc.addEventListener(
            "icegatheringstatechange",
            check
          );

          check();
        }
      );
    }

    /*
     * ------------------------------------------------------------
     * PEER CREATION
     * ------------------------------------------------------------
     */

    function getOrCreatePeer(
      remoteUserId: number
    ): Peer {
      const existing =
        peersRef.current.get(
          remoteUserId
        );

      if (
        existing &&
        existing.pc.connectionState !==
          "closed"
      ) {
        return existing;
      }

      if (existing) {
        try {
          existing.pc.close();
        } catch {}

        try {
          existing.audio.pause();
          existing.audio.srcObject =
            null;
          existing.audio.remove();
        } catch {}

        peersRef.current.delete(
          remoteUserId
        );
      }

      console.log(
        "VOICE: creating RTCPeerConnection",
        {
          remoteUserId,
          currentUserId,
        }
      );

      const pc =
        new RTCPeerConnection(
          RTC_CONFIG
        );

      const remoteStream =
        new MediaStream();

      const audio =
        document.createElement(
          "audio"
        );

      audio.autoplay = true;
      audio.controls = false;
      audio.volume = 1;

      audio.style.position =
        "fixed";
      audio.style.width = "1px";
      audio.style.height = "1px";
      audio.style.opacity = "0";
      audio.style.pointerEvents =
        "none";

      audio.srcObject =
        remoteStream;

      document.body.appendChild(
        audio
      );

      /*
       * ----------------------------------------------------------
       * LOCAL TRACKS
       * ----------------------------------------------------------
       */

      const localStream =
        localStreamRef.current;

      if (!localStream) {
        throw new Error(
          "Local microphone stream is missing"
        );
      }

      for (
        const track of
        localStream.getTracks()
      ) {
        console.log(
          "VOICE: add local track",
          {
            remoteUserId,
            kind: track.kind,
            id: track.id,
            enabled:
              track.enabled,
            readyState:
              track.readyState,
          }
        );

        pc.addTrack(
          track,
          localStream
        );
      }

      /*
       * ----------------------------------------------------------
       * REMOTE TRACK
       * ----------------------------------------------------------
       */

      pc.ontrack = async (
        event
      ) => {
        console.log(
          "VOICE: REMOTE TRACK",
          {
            remoteUserId,
            kind:
              event.track.kind,
            id:
              event.track.id,
            streams:
              event.streams.length,
          }
        );

        const alreadyAdded =
          remoteStream
            .getTracks()
            .some(
              (track) =>
                track.id ===
                event.track.id
            );

        if (!alreadyAdded) {
          remoteStream.addTrack(
            event.track
          );
        }

        audio.srcObject =
          remoteStream;

        try {
          await audio.play();

          console.log(
            "VOICE: remote audio playing",
            remoteUserId
          );
        } catch (error) {
          console.warn(
            "VOICE: audio.play() failed",
            {
              remoteUserId,
              error,
            }
          );
        }
      };

      /*
       * ----------------------------------------------------------
       * ICE CANDIDATE
       * ----------------------------------------------------------
       *
       * We are NOT sending candidates individually.
       *
       * We are collecting them into the SDP by waiting for ICE
       * gathering to complete.
       */

      pc.onicecandidate = (
        event
      ) => {
        if (
          event.candidate
        ) {
          console.log(
            "VOICE: LOCAL ICE CANDIDATE",
            {
              remoteUserId,
              candidate:
                event.candidate.toJSON(),
            }
          );
        } else {
          console.log(
            "VOICE: LOCAL ICE END",
            remoteUserId
          );
        }
      };

      /*
       * ----------------------------------------------------------
       * ICE GATHERING
       * ----------------------------------------------------------
       */

      pc.onicegatheringstatechange =
        () => {
          console.log(
            "VOICE: ICE GATHERING STATE",
            {
              remoteUserId,
              state:
                pc.iceGatheringState,
            }
          );
        };

      /*
       * ----------------------------------------------------------
       * ICE CONNECTION
       * ----------------------------------------------------------
       */

      pc.oniceconnectionstatechange =
        async () => {
          console.log(
            "VOICE: ICE CONNECTION STATE",
            {
              remoteUserId,
              state:
                pc.iceConnectionState,
            }
          );

          if (
            pc.iceConnectionState ===
              "connected" ||
            pc.iceConnectionState ===
              "completed"
          ) {
            console.log(
              "VOICE: ICE CONNECTED",
              remoteUserId
            );

            await dumpStats(
              pc,
              remoteUserId
            );
          }

          if (
            pc.iceConnectionState ===
            "failed"
          ) {
            console.error(
              "VOICE: ICE FAILED",
              remoteUserId
            );

            await dumpStats(
              pc,
              remoteUserId
            );
          }
        };

      /*
       * ----------------------------------------------------------
       * CONNECTION
       * ----------------------------------------------------------
       */

      pc.onconnectionstatechange =
        () => {
          console.log(
            "VOICE: CONNECTION STATE",
            {
              remoteUserId,
              state:
                pc.connectionState,
            }
          );

          if (
            pc.connectionState ===
            "connected"
          ) {
            console.log(
              "VOICE: PEER CONNECTED",
              remoteUserId
            );
          }

          if (
            pc.connectionState ===
            "failed"
          ) {
            console.error(
              "VOICE: PEER FAILED",
              remoteUserId
            );
          }
        };

      /*
       * ----------------------------------------------------------
       * SIGNALING
       * ----------------------------------------------------------
       */

      pc.onsignalingstatechange =
        () => {
          console.log(
            "VOICE: SIGNALING STATE",
            {
              remoteUserId,
              state:
                pc.signalingState,
            }
          );
        };

      /*
       * ----------------------------------------------------------
       * ICE ERRORS
       * ----------------------------------------------------------
       */

      pc.onicecandidateerror =
        (event) => {
          console.error(
            "VOICE: ICE CANDIDATE ERROR",
            {
              remoteUserId,
              errorCode:
                event.errorCode,
              errorText:
                event.errorText,
              url:
                event.url,
              address:
                event.address,
              port:
                event.port,
            }
          );
        };

      const peer: Peer = {
        pc,
        audio,
        remoteStream,
      };

      peersRef.current.set(
        remoteUserId,
        peer
      );

      console.log(
        "VOICE: PEER CREATED",
        {
          remoteUserId,
          signalingState:
            pc.signalingState,
          iceGatheringState:
            pc.iceGatheringState,
          iceConnectionState:
            pc.iceConnectionState,
          connectionState:
            pc.connectionState,
        }
      );

      return peer;
    }

    /*
     * ------------------------------------------------------------
     * OFFER
     * ------------------------------------------------------------
     */

    async function createOffer(
      remoteUserId: number
    ) {
      try {
        const peer =
          getOrCreatePeer(
            remoteUserId
          );

        const { pc } = peer;

        if (
          pc.signalingState !==
          "stable"
        ) {
          console.log(
            "VOICE: refusing offer; state isn't stable",
            {
              remoteUserId,
              state:
                pc.signalingState,
            }
          );

          return;
        }

        console.log(
          "VOICE: creating offer",
          remoteUserId
        );

        const offer =
          await pc.createOffer();

        console.log(
          "VOICE: OFFER CREATED",
          {
            remoteUserId,
            type: offer.type,
            sdpLength:
              offer.sdp?.length,
          }
        );

        await pc.setLocalDescription(
          offer
        );

        console.log(
          "VOICE: LOCAL OFFER SET",
          {
            remoteUserId,
            signalingState:
              pc.signalingState,
            iceGatheringState:
              pc.iceGatheringState,
            sdpLength:
              pc.localDescription
                ?.sdp.length,
          }
        );

        /*
         * Wait until candidates have been gathered.
         */
        await waitForIceGathering(
          pc,
          remoteUserId
        );

        if (cancelled) {
          return;
        }

        const sdp =
          pc.localDescription?.sdp;

        console.log(
          "VOICE: FINAL OFFER SDP",
          {
            remoteUserId,
            sdpLength:
              sdp?.length,
            hasCandidate:
              sdp?.includes(
                "a=candidate:"
              ),
            candidateCount:
              sdp
                ?.split("\n")
                .filter((line) =>
                  line.startsWith(
                    "a=candidate:"
                  )
                ).length ?? 0,
          }
        );

        if (!sdp) {
          console.error(
            "VOICE: no local offer SDP"
          );

          return;
        }

        sendSignal({
          type: "offer",
          from: currentUserId,
          to: remoteUserId,
          sdp,
        });
      } catch (error) {
        console.error(
          "VOICE: createOffer failed",
          {
            remoteUserId,
            error,
          }
        );
      }
    }

    /*
     * ------------------------------------------------------------
     * OFFER HANDLER
     * ------------------------------------------------------------
     */

    async function handleOffer(
      remoteUserId: number,
      sdp: string
    ) {
      try {
        const peer =
          getOrCreatePeer(
            remoteUserId
          );

        const { pc } = peer;

        console.log(
          "VOICE: RECEIVED OFFER",
          {
            remoteUserId,
            signalingState:
              pc.signalingState,
            sdpLength:
              sdp.length,
            hasCandidate:
              sdp.includes(
                "a=candidate:"
              ),
            candidateCount:
              sdp
                .split("\n")
                .filter((line) =>
                  line.startsWith(
                    "a=candidate:"
                  )
                ).length,
          }
        );

        if (
          pc.signalingState !==
          "stable"
        ) {
          console.warn(
            "VOICE: ignoring offer because peer isn't stable",
            {
              remoteUserId,
              state:
                pc.signalingState,
            }
          );

          return;
        }

        await pc.setRemoteDescription(
          {
            type: "offer",
            sdp,
          }
        );

        console.log(
          "VOICE: REMOTE OFFER SET",
          {
            remoteUserId,
            signalingState:
              pc.signalingState,
            iceConnectionState:
              pc.iceConnectionState,
          }
        );

        const answer =
          await pc.createAnswer();

        await pc.setLocalDescription(
          answer
        );

        console.log(
          "VOICE: LOCAL ANSWER SET",
          {
            remoteUserId,
            signalingState:
              pc.signalingState,
            iceGatheringState:
              pc.iceGatheringState,
          }
        );

        /*
         * Wait for answer ICE gathering too.
         */
        await waitForIceGathering(
          pc,
          remoteUserId
        );

        if (cancelled) {
          return;
        }

        const finalSdp =
          pc.localDescription?.sdp;

        console.log(
          "VOICE: FINAL ANSWER SDP",
          {
            remoteUserId,
            sdpLength:
              finalSdp?.length,
            hasCandidate:
              finalSdp?.includes(
                "a=candidate:"
              ),
            candidateCount:
              finalSdp
                ?.split("\n")
                .filter((line) =>
                  line.startsWith(
                    "a=candidate:"
                  )
                ).length ?? 0,
          }
        );

        if (!finalSdp) {
          console.error(
            "VOICE: no local answer SDP"
          );

          return;
        }

        sendSignal({
          type: "answer",
          from: currentUserId,
          to: remoteUserId,
          sdp: finalSdp,
        });

        console.log(
          "VOICE: ANSWER SENT",
          remoteUserId
        );
      } catch (error) {
        console.error(
          "VOICE: handleOffer failed",
          {
            remoteUserId,
            error,
          }
        );
      }
    }

    /*
     * ------------------------------------------------------------
     * ANSWER HANDLER
     * ------------------------------------------------------------
     */

    async function handleAnswer(
      remoteUserId: number,
      sdp: string
    ) {
      try {
        const peer =
          peersRef.current.get(
            remoteUserId
          );

        if (!peer) {
          console.error(
            "VOICE: answer received but peer doesn't exist",
            remoteUserId
          );

          return;
        }

        const { pc } = peer;

        console.log(
          "VOICE: RECEIVED ANSWER",
          {
            remoteUserId,
            signalingState:
              pc.signalingState,
            sdpLength:
              sdp.length,
            hasCandidate:
              sdp.includes(
                "a=candidate:"
              ),
            candidateCount:
              sdp
                .split("\n")
                .filter((line) =>
                  line.startsWith(
                    "a=candidate:"
                  )
                ).length,
          }
        );

        if (
          pc.signalingState !==
          "have-local-offer"
        ) {
          console.warn(
            "VOICE: ignoring answer; wrong signaling state",
            {
              remoteUserId,
              state:
                pc.signalingState,
            }
          );

          return;
        }

        await pc.setRemoteDescription(
          {
            type: "answer",
            sdp,
          }
        );

        console.log(
          "VOICE: REMOTE ANSWER SET",
          {
            remoteUserId,
            signalingState:
              pc.signalingState,
            iceConnectionState:
              pc.iceConnectionState,
          }
        );
      } catch (error) {
        console.error(
          "VOICE: handleAnswer failed",
          {
            remoteUserId,
            error,
          }
        );
      }
    }

    /*
     * ------------------------------------------------------------
     * USER LEFT
     * ------------------------------------------------------------
     */

    function removePeer(
      remoteUserId: number
    ) {
      const peer =
        peersRef.current.get(
          remoteUserId
        );

      if (!peer) {
        return;
      }

      console.log(
        "VOICE: REMOVING PEER",
        remoteUserId
      );

      peersRef.current.delete(
        remoteUserId
      );

      try {
        peer.pc.close();
      } catch {}

      try {
        peer.audio.pause();
        peer.audio.srcObject =
          null;
        peer.audio.remove();
      } catch {}
    }

    /*
     * ------------------------------------------------------------
     * STATS
     * ------------------------------------------------------------
     */

    async function dumpStats(
      pc: RTCPeerConnection,
      remoteUserId: number
    ) {
      try {
        const stats =
          await pc.getStats();

        const result: Record<
          string,
          unknown[]
        > = {
          candidatePairs: [],
          localCandidates: [],
          remoteCandidates: [],
          inbound: [],
          outbound: [],
          codecs: [],
        };

        stats.forEach(
          (report) => {
            if (
              report.type ===
              "candidate-pair"
            ) {
              result.candidatePairs.push(
                {
                  state:
                    report.state,
                  nominated:
                    report.nominated,
                  bytesSent:
                    report.bytesSent,
                  bytesReceived:
                    report.bytesReceived,
                  localCandidateId:
                    report.localCandidateId,
                  remoteCandidateId:
                    report.remoteCandidateId,
                }
              );
            }

            if (
              report.type ===
              "local-candidate"
            ) {
              result.localCandidates.push(
                {
                  candidateType:
                    report.candidateType,
                  protocol:
                    report.protocol,
                  address:
                    report.address,
                  port:
                    report.port,
                }
              );
            }

            if (
              report.type ===
              "remote-candidate"
            ) {
              result.remoteCandidates.push(
                {
                  candidateType:
                    report.candidateType,
                  protocol:
                    report.protocol,
                  address:
                    report.address,
                  port:
                    report.port,
                }
              );
            }

            if (
              report.type ===
                "inbound-rtp" &&
              report.kind ===
                "audio"
            ) {
              result.inbound.push(
                {
                  packetsReceived:
                    report.packetsReceived,
                  bytesReceived:
                    report.bytesReceived,
                  packetsLost:
                    report.packetsLost,
                  codecId:
                    report.codecId,
                }
              );
            }

            if (
              report.type ===
                "outbound-rtp" &&
              report.kind ===
                "audio"
            ) {
              result.outbound.push(
                {
                  packetsSent:
                    report.packetsSent,
                  bytesSent:
                    report.bytesSent,
                  codecId:
                    report.codecId,
                  ssrc:
                    report.ssrc,
                }
              );
            }

            if (
              report.type ===
                "codec" &&
              report.mimeType?.startsWith(
                "audio/"
              )
            ) {
              result.codecs.push(
                {
                  mimeType:
                    report.mimeType,
                  clockRate:
                    report.clockRate,
                  channels:
                    report.channels,
                  payloadType:
                    report.payloadType,
                }
              );
            }
          }
        );

        console.log(
          "VOICE: STATS",
          {
            remoteUserId,
            connectionState:
              pc.connectionState,
            iceConnectionState:
              pc.iceConnectionState,
            iceGatheringState:
              pc.iceGatheringState,
            signalingState:
              pc.signalingState,
            result,
          }
        );
      } catch (error) {
        console.error(
          "VOICE: stats failed",
          error
        );
      }
    }

    /*
     * ------------------------------------------------------------
     * START
     * ------------------------------------------------------------
     */

    async function start() {
      /*
       * Microphone.
       */

      try {
        const stream =
          await navigator.mediaDevices.getUserMedia(
            {
              audio: {
                echoCancellation: true,
                noiseSuppression: true,
                autoGainControl: true,
              },
              video: false,
            }
          );

        if (
          cancelled ||
          generation !==
            generationRef.current
        ) {
          stream
            .getTracks()
            .forEach((track) =>
              track.stop()
            );

          return;
        }

        localStreamRef.current =
          stream;

        console.log(
          "VOICE: MICROPHONE READY",
          stream
            .getTracks()
            .map(
              (track) => ({
                id: track.id,
                kind:
                  track.kind,
                enabled:
                  track.enabled,
                readyState:
                  track.readyState,
              })
            )
        );
      } catch (error) {
        console.error(
          "VOICE: MICROPHONE FAILED",
          error
        );

        return;
      }

      /*
       * WebSocket.
       */

      const ws =
        new WebSocket(
          `ws://localhost:8080/ws/voice/${channelId}`
        );

      wsRef.current =
        ws;

      ws.onopen = () => {
        if (
          generation !==
          generationRef.current
        ) {
          return;
        }

        console.log(
          "VOICE: WEBSOCKET OPEN",
          {
            channelId,
            currentUserId,
          }
        );

        flushOutgoingQueue();
      };

      ws.onerror = (
        event
      ) => {
        console.error(
          "VOICE: WEBSOCKET ERROR",
          event
        );
      };

      ws.onclose = (
        event
      ) => {
        console.log(
          "VOICE: WEBSOCKET CLOSED",
          {
            code:
              event.code,
            reason:
              event.reason,
            wasClean:
              event.wasClean,
          }
        );
      };

      ws.onmessage =
        async (event) => {
          if (
            cancelled ||
            generation !==
              generationRef.current
          ) {
            return;
          }

          try {
            const message =
              JSON.parse(
                event.data
              ) as SignalMessage;

            console.log(
              "VOICE: SIGNAL RECEIVE",
              message
            );

            switch (
              message.type
            ) {
              /*
               * ------------------------------------------------
               * USER JOINED
               * ------------------------------------------------
               */

              case "user_joined": {
                const remoteUserId =
                  getRemoteUserId(
                    message
                  );

                if (
                  remoteUserId ===
                    null ||
                  remoteUserId ===
                    currentUserId
                ) {
                  return;
                }

                /*
                 * Lower ID is always offerer.
                 */
                if (
                  currentUserId <
                  remoteUserId
                ) {
                  await createOffer(
                    remoteUserId
                  );
                }

                break;
              }

              /*
               * ------------------------------------------------
               * OFFER
               * ------------------------------------------------
               */

              case "offer": {
                const remoteUserId =
                  getRemoteUserId(
                    message
                  );

                if (
                  remoteUserId ===
                    null ||
                  !message.sdp
                ) {
                  console.error(
                    "VOICE: INVALID OFFER",
                    message
                  );

                  return;
                }

                await handleOffer(
                  remoteUserId,
                  message.sdp
                );

                break;
              }

              /*
               * ------------------------------------------------
               * ANSWER
               * ------------------------------------------------
               */

              case "answer": {
                const remoteUserId =
                  getRemoteUserId(
                    message
                  );

                if (
                  remoteUserId ===
                    null ||
                  !message.sdp
                ) {
                  console.error(
                    "VOICE: INVALID ANSWER",
                    message
                  );

                  return;
                }

                await handleAnswer(
                  remoteUserId,
                  message.sdp
                );

                break;
              }

              /*
               * ------------------------------------------------
               * ICE CANDIDATE
               * ------------------------------------------------
               *
               * This shouldn't be used by the current local
               * configuration because ICE is embedded in SDP.
               *
               * We still support it so the signaling protocol
               * remains compatible.
               */

              case "ice_candidate": {
                console.log(
                  "VOICE: IGNORING TRICKLE ICE",
                  message
                );

                break;
              }

              /*
               * ------------------------------------------------
               * USER LEFT
               * ------------------------------------------------
               */

              case "user_left": {
                const remoteUserId =
                  getRemoteUserId(
                    message
                  );

                if (
                  remoteUserId !==
                  null
                ) {
                  removePeer(
                    remoteUserId
                  );
                }

                break;
              }

              default: {
                console.log(
                  "VOICE: UNKNOWN SIGNAL",
                  message
                );
              }
            }
          } catch (error) {
            console.error(
              "VOICE: SIGNAL HANDLING FAILED",
              error
            );
          }
        };
    }

    start();

    /*
     * ------------------------------------------------------------
     * CLEANUP
     * ------------------------------------------------------------
     */

    return () => {
      if (
        generation !==
        generationRef.current
      ) {
        return;
      }

      console.log(
        "VOICE: CLEANUP",
        {
          generation,
          channelId,
          currentUserId,
        }
      );

      cancelled = true;

      cancelledRef.current =
        true;

      /*
       * Close peers.
       */

      for (
        const [
          remoteUserId,
          peer,
        ] of peersRef.current
      ) {
        console.log(
          "VOICE: CLOSING PEER",
          remoteUserId
        );

        try {
          peer.pc.ontrack =
            null;

          peer.pc.onicecandidate =
            null;

          peer.pc.onconnectionstatechange =
            null;

          peer.pc.oniceconnectionstatechange =
            null;

          peer.pc.close();
        } catch {}

        try {
          peer.audio.pause();
          peer.audio.srcObject =
            null;
          peer.audio.remove();
        } catch {}
      }

      peersRef.current.clear();

      /*
       * Stop microphone.
       */

      const stream =
        localStreamRef.current;

      if (stream) {
        stream
          .getTracks()
          .forEach((track) =>
            track.stop()
          );

        localStreamRef.current =
          null;
      }

      /*
       * Close WebSocket.
       */

      const ws =
        wsRef.current;

      if (ws) {
        try {
          ws.close();
        } catch {}

        wsRef.current =
          null;
      }

      outgoingQueueRef.current =
        [];
    };
  }, [
    channelId,
    currentUserId,
  ]);

  return null;
}
