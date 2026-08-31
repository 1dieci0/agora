import { useEffect } from "react";

export default function SfuTest() {
  useEffect(() => {
    let cancelled = false;

    let stream: MediaStream | null = null;
    let pc: RTCPeerConnection | null = null;
    let ws: WebSocket | null = null;

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

        /*
         * React may have already cleaned up this effect
         * while getUserMedia() was waiting.
         */
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

        const pc = new RTCPeerConnection({
          iceServers: [],
        });

        pc.ontrack = (event) => {
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

          const stream =
            event.streams[0] ??
            new MediaStream([event.track]);

          const audio =
            new Audio();

          audio.autoplay = true;
          audio.srcObject = stream;

          audio.play().catch((error) => {
            console.error(
              "SFU TEST: failed to play remote audio",
              error
            );
          });
        };

        /*
         * Add microphone.
         */

        for (const track of stream.getTracks()) {
          pc.addTrack(track, stream);
        }

        /*
         * ICE debugging.
         */

        pc.onicecandidate = (event) => {
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

        pc.onicegatheringstatechange = () => {
          console.log(
            "SFU TEST: ICE gathering state",
            pc?.iceGatheringState
          );
        };

        pc.oniceconnectionstatechange = () => {
          console.log(
            "SFU TEST: ICE connection state",
            pc?.iceConnectionState
          );
        };

        pc.onconnectionstatechange = () => {
          console.log(
            "SFU TEST: connection state",
            pc?.connectionState
          );
        };

        pc.onsignalingstatechange = () => {
          console.log(
            "SFU TEST: signaling state",
            pc?.signalingState
          );
        };

        /*
         * --------------------------------------------------------
         * WebSocket signaling
         * --------------------------------------------------------
         */
        const params = new URLSearchParams(window.location.search);

        const userId = params.get("user") ?? "1";


        const ws = new WebSocket(
          `ws://localhost:9000/sfu/1?user_id=${userId}`
        );

        ws.onopen = async () => {
          if (cancelled || !pc || !ws) {
            return;
          }

          console.log(
            "SFU TEST: signaling WebSocket open"
          );

          try {
            const offer =
              await pc.createOffer();

            if (cancelled) {
              return;
            }

            await pc.setLocalDescription(
              offer
            );

            /*
             * Non-trickle ICE for this first test.
             */

            if (
              pc.iceGatheringState !==
              "complete"
            ) {
              await new Promise<void>(
                (resolve) => {
                  const check = () => {
                    if (
                      pc?.iceGatheringState ===
                      "complete"
                    ) {
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

            if (
              cancelled ||
              !pc ||
              !ws
            ) {
              return;
            }

            const localDescription =
              pc.localDescription;

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

        ws.onmessage = async (event) => {
          try {
            const message = JSON.parse(event.data);

            console.log(
              "SFU TEST: signaling message",
              message
            );

            /*
            * --------------------------------------------------------
            * SFU -> Browser offer
            *
            * This happens when the SFU adds a remote user's
            * track to this PeerConnection.
            * --------------------------------------------------------
            */

            if (
              message.type === "offer" &&
              typeof message.sdp === "string"
            ) {
              console.log(
                "SFU TEST: received SFU offer"
              );

              await pc.setRemoteDescription({
                type: "offer",
                sdp: message.sdp,
              });

              console.log(
                "SFU TEST: SFU offer set"
              );

              const answer =
                await pc.createAnswer();

              await pc.setLocalDescription(
                answer
              );

              const localDescription =
                pc.localDescription;

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
              await pc.setRemoteDescription({
                type: "answer",
                sdp: message.sdp,
              });

              console.log(
                "SFU TEST: remote answer set"
              );

              return;
            }
          } catch (error) {
            console.error(
              "SFU TEST: failed handling signaling message",
              error
            );
          }
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
       * Close PeerConnection.
       */

      if (pc) {
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
  }, []);

  return null;
}
