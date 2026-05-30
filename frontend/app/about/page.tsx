import Navbar from "../../components/Navbar";

export default function About() {
  return (
    <>
      <Navbar />
      <main className="flex-1 flex items-center justify-center p-4">
        <div className="bg-surface border border-border rounded-md shadow p-8 max-w-md w-full flex flex-col gap-6">
          <div>
            <h1 className="text-2xl font-bold text-on-surface">SGame</h1>
            <p className="mt-1 text-sm text-on-surface-muted">
              A real-time multiplayer quiz game inspired by Jeopardy. Create
              rooms, pick question packs, and compete with friends.
            </p>
          </div>

          <div className="border-t border-border pt-6">
            <p className="text-xs font-semibold uppercase tracking-widest mb-3 text-on-surface-muted">
              Contact
            </p>
            <ul className="flex flex-col gap-2">
              <li>
                <a
                  className="flex items-center gap-2 text-sm text-primary underline underline-offset-2 hover:text-primary-hover"
                  href="mailto:holdennekt@gmail.com"
                >
                  Email
                </a>
              </li>
              <li>
                <a
                  className="flex items-center gap-2 text-sm text-primary underline underline-offset-2 hover:text-primary-hover"
                  href="https://t.me/holdennekt"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  Telegram
                </a>
              </li>
              <li>
                <span className="flex items-center gap-2 text-sm text-on-surface-muted">
                  Discord
                </span>
              </li>
            </ul>
          </div>
        </div>
      </main>
    </>
  );
}
