import { useState, type SyntheticEvent } from "react";
import { login } from "./api";
import type { User } from "./types";

type LoginProps = {
  onLogin: (user: User) => void;
  onSignup: () => void;
};

function Login({ onLogin, onSignup } : LoginProps) {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(event: SyntheticEvent<HTMLFormElement>) {
    event.preventDefault();

    setError("");
    setLoading(true);

    try {
      const result = await login(username, password);

      onLogin(result.user);
    } catch {
      setError("Invalid username or password");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="login-page">
      <form className="login-card" onSubmit={handleSubmit}>
        <h1>Welcome to Agora</h1>

        <p>Log in to continue</p>

        <input
          type="text"
          placeholder="Username"
          value={username}
          onChange={(event) => setUsername(event.target.value)}
        />

        <input
          type="password"
          placeholder="Password"
          value={password}
          onChange={(event) => setPassword(event.target.value)}
        />

        {error && <p className="error">{error}</p>}

        <button type="submit" disabled={loading}>
          {loading ? "Logging in..." : "Login"}
        </button>

        <p className="auth-switch">
          Don't have an account?{" "}
          <button
            type="button"
            onClick={onSignup}
          >
            Sign up
          </button>

        </p>
      </form>
    </div>
  );
}

export default Login;
