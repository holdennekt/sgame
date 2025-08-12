export default function SystemMessage({ text }: { text: string; }) {
  return (
    <div className="flex justify-center items-center">
      <div className="text-sm break-words rounded-xl py-2 px-4 background">
        <p>{text}</p>
      </div>
    </div>
  );
}
