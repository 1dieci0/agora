import { useEffect } from "react";
import { getVoiceToken } from "./api";

type VoiceConnectionProps = {
  channelId: number;
};

export default function VoiceConnection({
  channelId,
}: VoiceConnectionProps) {
  useEffect(() => {
    let cancelled = false;

    let stream: MediaStream | null = null;
    let pc: RTCPeerConnection | null = null;
    let ws: WebSocket | null = null;

    const remoteAudios: HTMLAudioElement[] = [];
    let negotiationChain = Promise.resolve();

    async function start() {
      try {
        /*
         * --------------------------------------------------------
         * Microphone
         * --------------------------------------------------------
         */

        stream =
          await navigator.mediaDevices.getUserMedia({
            audio: {
              echoCancellation: true,
              noiseSuppression: true,
              autoGainControl: true,
            },
            video: false,
          });

        if (cancelled) {
          stream
            .getTracks()
            .forEach((track) => track.stop());

          stream = null;
          return;
        }

        console.log(
          "SFU TEST: microphone ready"
        );

        /*
         * --------------------------------------------------------
         * PeerConnection
         * --------------------------------------------------------
         */

        const peerConnection = new RTCPeerConnection({
          iceServers: [],
        });

        pc = peerConnection;

        peerConnection.ontrack = (event) => {
          console.log(
            "SFU TEST: remote track received",
            {
              kind: event.track.kind,
              streams: event.streams.length,
            }
          );

          if (event.track.kind !== "audio") {
            return;
          }

          const remoteStream =
            event.streams[0] ??
            new MediaStream([event.track]);

          const audio = new Audio();

          audio.autoplay = true;
          audio.srcObject = remoteStream;

          remoteAudios.push(audio);

          audio.play().catch((error) => {
            if (!cancelled) {
              console.error(
                "SFU TEST: failed to play remote audio",
                error
              );
            }
          });

          /*
           * If the SFU/browser ends this track,
           * make sure the audio element stops.
           */
          event.track.onended = () => {
            audio.pause();
            audio.srcObject = null;

            const index =
              remoteAudios.indexOf(audio);

            if (index !== -1) {
              remoteAudios.splice(index, 1);
            }
          };
        };

        /*
         * Add microphone.
         */

        for (const track of stream.getTracks()) {
          peerConnection.addTrack(track, stream);
        }

        /*
         * ICE debugging.
         */

        peerConnection.onicecandidate = (event) => {
          if (event.candidate) {
            console.log(
              "SFU TEST: local ICE candidate",
              event.candidate
            );
          } else {
            console.log(
              "SFU TEST: ICE gathering complete"
            );
          }
        };

        peerConnection.onicegatheringstatechange = () => {
          console.log(
            "SFU TEST: ICE gathering state",
            peerConnection?.iceGatheringState
          );
        };

        peerConnection.oniceconnectionstatechange = () => {
          console.log(
            "SFU TEST: ICE connection state",
            peerConnection?.iceConnectionState
          );
        };

        peerConnection.onconnectionstatechange = () => {
          console.log(
            "SFU TEST: connection state",
            peerConnection?.connectionState
          );
        };

        peerConnection.onsignalingstatechange = () => {
          console.log(
            "SFU TEST: signaling state",
            peerConnection?.signalingState
          );
        };

        /*
         * --------------------------------------------------------
         * Get SFU token
         * --------------------------------------------------------
         */

        const token =
          await getVoiceToken(channelId);

        if (cancelled) {
          return;
        }

        /*
         * --------------------------------------------------------
         * WebSocket signaling
         * --------------------------------------------------------
         */

        ws = new WebSocket(
          `ws://localhost:9000/sfu/${channelId}?token=${encodeURIComponent(token)}`,
        );

        ws.onopen = async () => {
          if (cancelled || !peerConnection || !ws) {
            return;
          }

          console.log(
            "SFU TEST: signaling WebSocket open"
          );

          try {
            const offer =
              await peerConnection.createOffer();

            if (cancelled) {
              return;
            }

            await peerConnection.setLocalDescription(
              offer
            );

            /*
             * Non-trickle ICE for this first test.
             */

            if (
              peerConnection.iceGatheringState !==
              "complete"
            ) {
              await new Promise<void>(
                (resolve) => {
                  const check = () => {
                    if (
                      peerConnection?.iceGatheringState ===
                      "complete"
                    ) {
                      peerConnection.removeEventListener(
                        "icegatheringstatechange",
                        check
                      );

                      resolve();
                    }
                  };

                  peerConnection.addEventListener(
                    "icegatheringstatechange",
                    check
                  );

                  check();
                }
              );
            }

            if (
              cancelled ||
              !peerConnection ||
              !ws
            ) {
              return;
            }

            const localDescription =
              peerConnection.localDescription;

            if (!localDescription) {
              throw new Error(
                "SFU TEST: local description is missing"
              );
            }

            console.log(
              "SFU TEST: sending offer",
              {
                sdpLength:
                  localDescription.sdp.length,

                hasCandidate:
                  localDescription.sdp.includes(
                    "a=candidate:"
                  ),
              }
            );

            ws.send(
              JSON.stringify({
                type: "offer",
                sdp: localDescription.sdp,
              })
            );
          } catch (error) {
            if (!cancelled) {
              console.error(
                "SFU TEST: offer failed",
                error
              );
            }
          }
        };

        ws.onmessage = (event) => {
          negotiationChain = negotiationChain
            .then(async () => {
              if (cancelled || !peerConnection || !ws) {
                return;
              }

              const message = JSON.parse(event.data);

              console.log(
                "SFU TEST: signaling message",
                message
              );

              /*
              * --------------------------------------------------------
              * SFU -> Browser offer
              *
              * Serialize these so multiple SFU renegotiations
              * cannot modify the PeerConnection simultaneously.
              * --------------------------------------------------------
              */

              if (
                message.type === "offer" &&
                typeof message.sdp === "string"
              ) {
                console.log(
                  "SFU TEST: received SFU offer"
                );

                await peerConnection.setRemoteDescription({
                  type: "offer",
                  sdp: message.sdp,
                });

                console.log(
                  "SFU TEST: SFU offer set"
                );

                const answer =
                  await peerConnection.createAnswer();

                await peerConnection.setLocalDescription(
                  answer
                );

                const localDescription =
                  peerConnection.localDescription;

                if (!localDescription) {
                  throw new Error(
                    "SFU TEST: local description missing after renegotiation"
                  );
                }

                console.log(
                  "SFU TEST: sending renegotiation answer"
                );

                ws.send(
                  JSON.stringify({
                    type: "answer",
                    sdp: localDescription.sdp,
                  })
                );

                return;
              }

              /*
              * --------------------------------------------------------
              * SFU -> Browser initial answer
              * --------------------------------------------------------
              */

              if (
                message.type === "answer" &&
                typeof message.sdp === "string"
              ) {
                await peerConnection.setRemoteDescription({
                  type: "answer",
                  sdp: message.sdp,
                });

                console.log(
                  "SFU TEST: remote answer set"
                );
              }
            })
            .catch((error) => {
              if (!cancelled) {
                console.error(
                  "SFU TEST: failed handling signaling message",
                  error
                );
              }
            });
        };

        ws.onerror = (event) => {
          if (!cancelled) {
            console.error(
              "SFU TEST: WebSocket error",
              event
            );
          }
        };

        ws.onclose = (event) => {
          if (!cancelled) {
            console.log(
              "SFU TEST: WebSocket closed",
              {
                code: event.code,
                reason: event.reason,
              }
            );
          }
        };
      } catch (error) {
        if (!cancelled) {
          console.error(
            "SFU TEST: startup failed",
            error
          );
        }
      }
    }

    start();

    /*
     * --------------------------------------------------------
     * Cleanup
     * --------------------------------------------------------
     */

    return () => {
      cancelled = true;

      /*
       * Stop microphone.
       */

      if (stream) {
        stream
          .getTracks()
          .forEach((track) => track.stop());

        stream = null;
      }

      /*
       * Stop remote audio.
       */

      for (const audio of remoteAudios) {
        audio.pause();
        audio.srcObject = null;
      }

      remoteAudios.length = 0;

      /*
       * Close PeerConnection.
       */

      if (pc) {
        pc.ontrack = null;
        pc.onicecandidate = null;
        pc.onicegatheringstatechange = null;
        pc.oniceconnectionstatechange = null;
        pc.onconnectionstatechange = null;
        pc.onsignalingstatechange = null;

        try {
          pc.close();
        } catch {}

        pc = null;
      }

      /*
       * Close WebSocket.
       */

      if (ws) {
        ws.onopen = null;
        ws.onmessage = null;
        ws.onerror = null;
        ws.onclose = null;

        try {
          ws.close();
        } catch {}

        ws = null;
      }
    };
  }, [channelId]);

  return null;
}
