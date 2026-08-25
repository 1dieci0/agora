import { useEffect, useState } from "react";
import Login from "./Login";
import Chat from "./Chat";
import { getMe } from "./api";
import type { User } from "./types";

function App() {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

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

  if (loading) {
    return <p>Loading...</p>;
  }

  if (!user) {
    return <Login onLogin={setUser} />;
  }

  return <Chat user={user} />;
}

export default App;