import { useState, type FormEvent } from "react";

type CreateChannelModalProps = {
  onClose: () => void;
  onCreate: (
    name: string,
    type: "text" | "voice",
  ) => Promise<void>;
};

function CreateChannelModal({
  onClose,
  onCreate,
}: CreateChannelModalProps) {
  const [name, setName] = useState("");
  const [type, setType] = useState<"text" | "voice">("text");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const trimmedName = name.trim();

    if (!trimmedName) {
      setError("Channel name is required");
      return;
    }

    if (trimmedName.length > 50) {
      setError("Channel name is too long");
      return;
    }

    setLoading(true);
    setError("");

    try {
      await onCreate(trimmedName, type);
      onClose();
    } catch (error) {
      console.error(error);
      setError("Could not create channel");
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
          <h2>Create a channel</h2>

          <button
            type="button"
            onClick={onClose}
            disabled={loading}
          >
            ×
          </button>
        </div>

        <p className="modal-description">
          Add a new channel to your server.
        </p>

        <form onSubmit={handleSubmit}>
          <label htmlFor="channel-name">
            Channel name
          </label>

          <input
            id="channel-name"
            type="text"
            value={name}
            onChange={(event) => setName(event.target.value)}
            placeholder="general"
            maxLength={50}
            autoFocus
            disabled={loading}
          />

          <label htmlFor="channel-type">
            Channel type
          </label>

          <select
            id="channel-type"
            value={type}
            onChange={(event) =>
              setType(event.target.value as "text" | "voice")
            }
            disabled={loading}
          >
            <option value="text">Text</option>
            <option value="voice">Voice</option>
          </select>

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
              {loading ? "Creating..." : "Create Channel"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

export default CreateChannelModal;
