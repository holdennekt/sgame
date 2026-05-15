export default function TextPanel({
  topText,
  bottomText,
}: {
  topText: string;
  bottomText?: string | null;
}) {
  return (
    <div className="w-full h-full flex flex-col justify-center items-center gap-2 p-10">
      <p className="text-center text-3xl font-semibold">{topText}</p>
      {bottomText && (
        <p className="text-center text-2xl font-normal">{bottomText}</p>
      )}
    </div>
  );
}
