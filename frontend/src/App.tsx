import { useEffect, useState } from "react";
import Login from "./Login";
import Chat from "./Chat";
import { getMe, logout} from "./api";
import type { User } from "./types";
import Signup from "./Signup";

function App() {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [showSignup, setShowSignup] = useState(false);

  useEffect(() => {
    async function checkAuth() {
      try {
        const user = await getMe();
        setUser(user);
      } catch {
        setUser(null);
      } finally {
        setLoading(false);
      }
    }

    checkAuth();
  }, []);

  async function handleLogout() {
    try {
      await logout();
      setUser(null);
    } catch (error) {
      console.error("Logout failed:", error);
    }
  }

  if (loading) {
    return <p>Loading...</p>;
  }

  if (!user) {
    if (showSignup) {
      return (
        <Signup
          onSignup={setUser}
          onLogin={() => setShowSignup(false)}
        />
      );
    }

    return (
      <Login
        onLogin={setUser}
        onSignup={() => setShowSignup(true)}
      />
    );
  }

  return <Chat
    user={user}
    onLogout={handleLogout}
  />
}

export default App;
