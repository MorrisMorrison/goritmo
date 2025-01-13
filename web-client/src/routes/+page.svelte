<script>
  import { goto } from '$app/navigation'; 

  async function createRoom() {
    try {
      const response = await fetch('http://localhost:8080/rooms', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        }
      });

      if (!response.ok) {
        throw new Error('Failed to create room');
      }

      const data = await response.json();
      const roomID = data.roomID;

      goto(`/room/${roomID}`);
    } catch (error) {
      console.error('Error creating room:', error);
    }
  }
</script>

<button on:click={createRoom}>Create Room</button>
