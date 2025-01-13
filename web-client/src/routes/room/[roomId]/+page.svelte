<script lang="ts">
  export let roomId: string;
  let localStream: MediaStream | null = null;
  let remoteStream: MediaStream | null = null;
  let localVideo: HTMLVideoElement | null = null;
  let remoteVideo: HTMLVideoElement | null = null;
  let peerConnection: RTCPeerConnection | null = null;

  const signalingServer = new WebSocket("ws://localhost:8080/ws?room=" + roomId);

  const configuration: RTCConfiguration = {
    iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
  };

  signalingServer.onmessage = async (message) => {
    const data = JSON.parse(message.data);
    switch (data.type) {
      case "offer":
        await handleOffer(data.offer);
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

  function setupPeerConnection() {
    if (!peerConnection){
      console.error("No peer connection")
      return;
    }

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
  }

  async function startStream() {
    console.log("startStream()");
    try {
      localStream = await navigator.mediaDevices.getUserMedia({ video: false, audio: true });
      if (localVideo && localStream) {
        localVideo.srcObject = localStream;
      }

      peerConnection = new RTCPeerConnection(configuration);
      setupPeerConnection();

      localStream.getTracks().forEach((track) => {
        peerConnection?.addTrack(track, localStream as MediaStream);
      });
    } catch (error) {
      console.error("Error starting stream:", error);
    }
  }

  async function startScreenShare() {
    try {
      const screenStream = await navigator.mediaDevices.getDisplayMedia({
        video: true,
        audio: false,
      });

      if (localVideo) {
        localVideo.srcObject = screenStream;
      }

      screenStream.getTracks().forEach((track) => {
        peerConnection?.addTrack(track, screenStream);
      });
    } catch (error) {
      console.error("Error capturing screen:", error);
    }
  }

  async function createOffer() {
    console.log("createOffer()");
    if (peerConnection) {
      const offer = await peerConnection.createOffer();
      await peerConnection.setLocalDescription(offer);
      signalingServer.send(JSON.stringify({ type: "offer", offer }));
    }
  }

  async function handleOffer(offer: RTCSessionDescriptionInit) {
    console.log("handleOffer()");
    if (!peerConnection) {
      peerConnection = new RTCPeerConnection(configuration);
      setupPeerConnection();
    }

    await peerConnection.setRemoteDescription(new RTCSessionDescription(offer));
    const answer = await peerConnection.createAnswer();
    await peerConnection.setLocalDescription(answer);
    signalingServer.send(JSON.stringify({ type: "answer", answer }));
  }

  async function startViewer() {
    console.log("startViewer()");
    if (!peerConnection) {
      peerConnection = new RTCPeerConnection(configuration);
      setupPeerConnection();
    }

    signalingServer.send(JSON.stringify({ type: "viewer-join" }));
  }
</script>

<main>
  <h1>WebRTC + WebSocket with Svelte (TypeScript)</h1>
  <h2>Welcome to Room {roomId}</h2>
  <div>
    <!-- svelte-ignore a11y_media_has_caption -->
    <video autoplay bind:this={localVideo} playsinline></video>
    
    <!-- svelte-ignore a11y_media_has_caption -->
    <video autoplay bind:this={remoteVideo} playsinline></video>
  </div>
  <button on:click={startStream}>Start Stream</button>
  <button on:click={startScreenShare}>Start Screenshare</button>
  <button on:click={createOffer}>Create Offer</button>
  <button on:click={startViewer}>Join as Viewer</button>
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
