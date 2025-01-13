// src/routes/room/[roomId]/+page.ts
import type { PageLoad } from './$types';

export const load: PageLoad = ({ params }) => {
  return {
    roomId: params.roomId
  };
};
