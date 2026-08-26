import { useState, type FormEvent } from "react";

type JoinServerModalProps = {
  onClose: () => void;
  onJoin: (code: string) => Promise<void>;
};

function JoinServerModal({
  onClose,
  onJoin,
}: JoinServerModalProps) {
  const [code, setCode] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const trimmedCode = code.trim();

    if (!trimmedCode) {
      setError("Invite code is required");
      return;
    }

    setLoading(true);
    setError("");

    try {
      await onJoin(trimmedCode);
      onClose();
    } catch (error) {
      console.error(error);
      setError("Invalid or expired invite");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="modal-backdrop" onMouseDown={onClose}>
      <div
        className="modal"
        onMouseDown={(event) => event.stopPropagation()}
      >
        <div className="modal-header">
          <h2>Join a server</h2>

          <button
            type="button"
            onClick={onClose}
            disabled={loading}
          >
            ×
          </button>
        </div>

        <p className="modal-description">
          Enter an invite code to join a server.
        </p>

        <form onSubmit={handleSubmit}>
          <label htmlFor="invite-code">
            Invite code
          </label>

          <input
            id="invite-code"
            type="text"
            value={code}
            onChange={(event) => setCode(event.target.value)}
            placeholder="Enter invite code"
            autoFocus
            disabled={loading}
          />

          {error && (
            <p className="modal-error">
              {error}
            </p>
          )}

          <div className="modal-actions">
            <button
              type="button"
              onClick={onClose}
              disabled={loading}
            >
              Cancel
            </button>

            <button
              type="submit"
              disabled={loading || !code.trim()}
            >
              {loading ? "Joining..." : "Join Server"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

export default JoinServerModal;

