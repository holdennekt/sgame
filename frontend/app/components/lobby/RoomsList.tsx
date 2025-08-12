import RoomLobbyPage, { RoomLobby } from "./Room";

export default function RoomsList({
  rooms,
  openPasswordModal,
}: {
  rooms: RoomLobby[];
  openPasswordModal: (roomId: string) => void;
}) {
  return (
    <div
      className={`flex flex-col gap-2 flex-auto min-w-0 min-h-0 mt-3 
      overflow-x-clip overflow-y-auto rounded surface`}
    >
      {rooms.length ? (
        rooms.map((room, index) => (
          <RoomLobbyPage
            key={index}
            room={room}
            openPasswordModal={openPasswordModal}
          />
        ))
      ) : (
        <div className="w-full h-full flex justify-center items-center">
          <p className="text-lg">No rooms yet</p>
        </div>
      )}
    </div>
  );
}
