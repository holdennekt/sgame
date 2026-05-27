export default function SystemMessage({ text }: { text: string }) {
  return (
    <div className="flex justify-center items-center my-1">
      <p className="text-xs text-on-surface-muted bg-surface-raised px-3 py-1 rounded-full">
        {text}
      </p>
    </div>
  );
}
