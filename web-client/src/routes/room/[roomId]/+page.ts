import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
  const rooms = await fetchRooms();

  return {
    roomId: params.roomId,
    rooms
  };
};

export type Room = {
  id: string;
  peerCount: number;
}

async function fetchRooms(): Promise<Room[]> {
  const response = await fetch('http://localhost:8080/rooms');
  if (!response.ok) {
    throw new Error('Failed to fetch rooms');
  }

  const rooms: Room[] = await response.json();
  console.log(rooms)
  return rooms;
}
