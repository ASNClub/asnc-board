import { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { userManager, loadSessionSync, sessionFromOidcUser, IDP } from './lib/auth';
import { setAccessToken, setOnAuthExpired } from './lib/api';

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [auth, setAuth] = useState(() => {
    const s = loadSessionSync();
    const tok = s.status === 'authenticated' ? s.accessToken : null;
    setAccessToken(tok);
    return s;
  });

  useEffect(() => {
    const tok = auth.status === 'authenticated' ? auth.accessToken : null;
    setAccessToken(tok);
  }, [auth]);

  useEffect(() => {
    if (!userManager) return;

    userManager.getUser().then(user => {
      if (user && !user.expired) setAuth(sessionFromOidcUser(user));
      else setAuth({ status: 'guest' });
    });

    const onLoaded    = (user) => setAuth(sessionFromOidcUser(user));
    const onSignedOut = ()     => setAuth({ status: 'guest' });

    userManager.events.addUserLoaded(onLoaded);
    userManager.events.addUserSignedOut(onSignedOut);

    setOnAuthExpired(() => {
      userManager.removeUser().catch(() => {});
      setAuth({ status: 'guest' });
    });

    return () => {
      userManager.events.removeUserLoaded(onLoaded);
      userManager.events.removeUserSignedOut(onSignedOut);
      setOnAuthExpired(null);
    };
  }, []);

  const login = useCallback(() =>
    userManager?.signinRedirect(), []);

  const loginWith = useCallback((provider) => {
    const idpId = IDP[provider];
    return userManager?.signinRedirect(
      idpId ? { extraQueryParams: { idp_hint: idpId } } : undefined
    );
  }, []);

  const logout = useCallback(() =>
    userManager?.signoutRedirect(), []);

  return (
    <AuthContext.Provider value={{
      auth,
      isAuthenticated: auth.status === 'authenticated',
      login,
      loginWith,
      logout,
    }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used inside AuthProvider');
  return ctx;
}
