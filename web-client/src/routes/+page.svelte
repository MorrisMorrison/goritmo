<script lang="ts">
    let localStream: MediaStream | null = null;
    let remoteStream: MediaStream | null = null;
    let localVideo: HTMLVideoElement | null = null;
    let remoteVideo: HTMLVideoElement | null = null;
    let peerConnection: RTCPeerConnection | null = null;
  
    const signalingServer = new WebSocket("ws://localhost:8080/ws"); // Adjust URL as needed
  
    const configuration: RTCConfiguration = {
      iceServers: [{ urls: "stun:stun.l.google.com:19302" }], // Public STUN server
    };
  
    async function startStream() {
      try {
        localStream = await navigator.mediaDevices.getUserMedia({ video: true, audio: true });
        if (localVideo && localStream) {
          localVideo.srcObject = localStream;
        }
  
        peerConnection = new RTCPeerConnection(configuration);
  
        localStream?.getTracks().forEach((track) => {
          peerConnection?.addTrack(track, localStream as MediaStream);
        });
  
        peerConnection.ontrack = (event) => {
          remoteStream = event.streams[0];
          if (remoteVideo && remoteStream) {
            remoteVideo.srcObject = remoteStream;
          }
        };
  
        peerConnection.onicecandidate = (event) => {
          if (event.candidate) {
            signalingServer.send(JSON.stringify({ type: "ice-candidate", candidate: event.candidate }));
          }
        };
  
        signalingServer.onmessage = async (message) => {
          const data = JSON.parse(message.data);
          switch (data.type) {
            case "offer":
              if (peerConnection) {
                await peerConnection.setRemoteDescription(new RTCSessionDescription(data.offer));
                const answer = await peerConnection.createAnswer();
                await peerConnection.setLocalDescription(answer);
                signalingServer.send(JSON.stringify({ type: "answer", answer }));
              }
              break;
  
            case "answer":
              if (peerConnection) {
                await peerConnection.setRemoteDescription(new RTCSessionDescription(data.answer));
              }
              break;
  
            case "ice-candidate":
              if (peerConnection && data.candidate) {
                await peerConnection.addIceCandidate(new RTCIceCandidate(data.candidate));
              }
              break;
          }
        };
      } catch (error) {
        console.error("Error starting stream:", error);
      }
    }
  
    async function createOffer() {
      if (peerConnection) {
        const offer = await peerConnection.createOffer();
        await peerConnection.setLocalDescription(offer);
        signalingServer.send(JSON.stringify({ type: "offer", offer }));
      }
    }
  </script>
  
  <main>
    <h1>WebRTC + WebSocket with Svelte (TypeScript)</h1>
    <div>
      <video autoplay bind:this={localVideo} playsinline></video>
      <video autoplay bind:this={remoteVideo} playsinline></video>
    </div>
    <button on:click={startStream}>Start Stream</button>
    <button on:click={createOffer}>Create Offer</button>
  </main>
  
  <style>
    video {
      width: 45%;
      margin: 10px;
      border: 1px solid #ccc;
    }
  
    button {
      margin: 10px;
      padding: 10px;
      font-size: 16px;
      cursor: pointer;
    }
  </style>
  