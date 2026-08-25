import { useState, type FormEvent } from "react";

type CreateServerModalProps = {
  onClose: () => void;
  onCreate: (name: string) => Promise<void>;
};

function CreateServerModal({
  onClose,
  onCreate,
}: CreateServerModalProps) {
  const [name, setName] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const trimmedName = name.trim();

    if (!trimmedName) {
      setError("Server name is required");
      return;
    }

    if (trimmedName.length > 50) {
      setError("Server name is too long");
      return;
    }

    setLoading(true);
    setError("");

    try {
      await onCreate(trimmedName);
      onClose();
    } catch (error) {
      console.error(error);
      setError("Could not create server");
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
          <h2>Create a server</h2>

          <button
            type="button"
            onClick={onClose}
            disabled={loading}
          >
            ×
          </button>
        </div>

        <p className="modal-description">
          Give your new server a name.
        </p>

        <form onSubmit={handleSubmit}>
          <label htmlFor="server-name">
            Server name
          </label>

          <input
            id="server-name"
            type="text"
            value={name}
            onChange={(event) => setName(event.target.value)}
            placeholder="My server"
            maxLength={50}
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
              disabled={loading || !name.trim()}
            >
              {loading ? "Creating..." : "Create Server"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

export default CreateServerModal;

