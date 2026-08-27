
import { useState, type SyntheticEvent } from "react";
import { register } from "./api";
import type { User } from "./types";

type SignupProps = {
  onSignup: (user: User) => void;
  onLogin: () => void;
};

function Signup({ onSignup, onLogin }: SignupProps) {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(event: SyntheticEvent<HTMLFormElement>) {
    event.preventDefault();

    setError("");

    if (username.length < 3) {
      setError("Username must be at least 3 characters");
      return;
    }

    if (username.length > 20) {
      setError("Username must be at most 20 characters");
      return;
    }

    if (password.length < 3) {
      setError("Password must be at least 3 characters");
      return;
    }

    if (password.length > 20) {
      setError("Password must be at most 20 characters");
      return;
    }

    setLoading(true);

    try {
      const user = await register(username, password);

      onSignup(user);
    } catch (error) {
      if (error instanceof Error) {
        setError(error.message);
      } else {
        setError("Could not create account");
      }
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="auth-page">
      <div className="auth-card">
        <h1>Create an account</h1>

        <p className="auth-subtitle">
          Join Agora and start chatting.
        </p>

        <form onSubmit={handleSubmit}>
          <label htmlFor="username">
            Username
          </label>

          <input
            id="username"
            type="text"
            value={username}
            onChange={(event) =>
              setUsername(event.target.value)
            }
            placeholder="Username"
            autoComplete="username"
            disabled={loading}
          />

          <label htmlFor="password">
            Password
          </label>

          <input
            id="password"
            type="password"
            value={password}
            onChange={(event) =>
              setPassword(event.target.value)
            }
            placeholder="Password"
            autoComplete="new-password"
            disabled={loading}
          />

          {error && (
            <p className="auth-error">
              {error}
            </p>
          )}

          <button
            type="submit"
            disabled={
              loading ||
              !username ||
              !password
            }
          >
            {loading ? "Creating account..." : "Sign up"}
          </button>
        </form>

        <p className="auth-switch">
          Already have an account?{" "}
          <button
            type="button"
            onClick={onLogin}
            disabled={loading}
          >
            Log in
          </button>
        </p>
      </div>
    </div>
  );
}

export default Signup;
