<script lang="ts">
    let localStream: MediaStream | null = null;
    let remoteStream: MediaStream | null = null;
    let localVideo: HTMLVideoElement | null = null;
    let remoteVideo: HTMLVideoElement | null = null;
    let peerConnection: RTCPeerConnection | null = null;
  
    const signalingServer = new WebSocket("ws://localhost:8080/ws?room=1");
  
    const configuration: RTCConfiguration = {
      iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
    };

    async function startScreenShare() {
  try {
    const screenStream = await navigator.mediaDevices.getDisplayMedia({
      video: true,
      audio: false, // Enable audio if you want to capture system audio
    });
    
    // Display the screen stream locally
    if (localVideo) {
      localVideo.srcObject = screenStream;
    }

    // Add screen stream tracks to peer connection
    screenStream.getTracks().forEach((track) => {
      peerConnection?.addTrack(track, screenStream);
    });
  } catch (error) {
    console.error('Error capturing screen:', error);
  }
}

  
    async function startStream() {
      console.log("startStream()");

      try {
        localStream = await navigator.mediaDevices.getUserMedia({ video: false, audio: true });
        if (localVideo && localStream) {
          localVideo.srcObject = localStream;
        }
  
        peerConnection = new RTCPeerConnection(configuration);
  
        localStream?.getTracks().forEach((track) => {
          peerConnection?.addTrack(track, localStream as MediaStream);
        });
  
        peerConnection.ontrack = (event) => {
          console.log("ontrack: " + event)
          remoteStream = event.streams[0];
          if (remoteVideo && remoteStream) {
            remoteVideo.srcObject = remoteStream;
          }
        };
  
        peerConnection.onicecandidate = (event) => {
          console.log("onicecandidate: " + event)
          if (event.candidate) {
            signalingServer.send(JSON.stringify({ type: "ice-candidate", candidate: event.candidate }));
          }
        };
  
        signalingServer.onmessage = async (message) => {
          console.log("onmessage: " + event)
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
      console.log("createOffer()")

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

      <!-- svelte-ignore a11y_media_has_caption -->
      <video autoplay bind:this={localVideo} playsinline></video>
      
      <!-- svelte-ignore a11y_media_has_caption -->
      <video autoplay bind:this={remoteVideo} playsinline></video>
    </div>
    <button on:click={startStream}>Start Stream</button>
    <button on:click={startScreenShare}>Start Screenshare</button>
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
  