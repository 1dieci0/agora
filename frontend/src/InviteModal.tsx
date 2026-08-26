import { useState } from "react";

type InviteModalProps = {
  serverID: number;
  onClose: () => void;
  onCreateInvite: (serverID: number) => Promise<string>;
};

function InviteModal({
  serverID,
  onClose,
  onCreateInvite,
}: InviteModalProps) {
  const [code, setCode] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleCreateInvite() {
    setLoading(true);
    setError("");

    try {
      const inviteCode = await onCreateInvite(serverID);
      setCode(inviteCode);
    } catch (error) {
      console.error(error);
      setError("Could not create invite");
    } finally {
      setLoading(false);
    }
  }

  async function handleCopy() {
    if (!code) {
      return;
    }

    await navigator.clipboard.writeText(code);
  }

  return (
    <div className="modal-backdrop" onMouseDown={onClose}>
      <div
        className="modal"
        onMouseDown={(event) => event.stopPropagation()}
      >
        <div className="modal-header">
          <h2>Invite people</h2>

          <button
            type="button"
            onClick={onClose}
          >
            ×
          </button>
        </div>

        <p className="modal-description">
          Create an invite code and send it to someone.
        </p>

        {code ? (
          <>
            <input
              value={code}
              readOnly
            />

            <div className="modal-actions">
              <button
                type="button"
                onClick={handleCopy}
              >
                Copy
              </button>

              <button
                type="button"
                onClick={onClose}
              >
                Done
              </button>
            </div>
          </>
        ) : (
          <>
            {error && (
              <p className="modal-error">
                {error}
              </p>
            )}

            <div className="modal-actions">
              <button
                type="button"
                onClick={onClose}
              >
                Cancel
              </button>

              <button
                type="button"
                onClick={handleCreateInvite}
                disabled={loading}
              >
                {loading ? "Creating..." : "Create Invite"}
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  );
}

export default InviteModal;
