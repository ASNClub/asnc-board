import { UserManager, WebStorageStateStore } from 'oidc-client-ts';

const DEV_AUTH_ID   = import.meta.env.VITE_DEV_AUTH_ID;
const ZITADEL_URL   = import.meta.env.VITE_ZITADEL_URL;
const CLIENT_ID     = import.meta.env.VITE_ZITADEL_CLIENT_ID;

export const IDP = {
  github: import.meta.env.VITE_ZITADEL_IDP_GITHUB,
};

// UserManager создаётся только когда есть реальный Zitadel (не dev-bypass)
export const userManager =
  !DEV_AUTH_ID && ZITADEL_URL && CLIENT_ID
    ? new UserManager({
        authority:                 ZITADEL_URL,
        client_id:                 CLIENT_ID,
        redirect_uri:              `${window.location.origin}/callback`,
        post_logout_redirect_uri:  window.location.origin,
        scope:                     'openid profile email',
        userStore:                 new WebStorageStateStore({ store: localStorage }),
      })
    : null;

// Синхронное чтение сессии из localStorage (без await — для инициализации)
export function loadSessionSync() {
  if (DEV_AUTH_ID) {
    return { status: 'authenticated', accessToken: '', sub: DEV_AUTH_ID };
  }
  if (!ZITADEL_URL || !CLIENT_ID) return { status: 'guest' };

  const key = `oidc.user:${ZITADEL_URL}:${CLIENT_ID}`;
  try {
    const raw = localStorage.getItem(key);
    if (!raw) return { status: 'guest' };
    const data = JSON.parse(raw);
    if (!data.access_token) return { status: 'guest' };
    if (data.expires_at && data.expires_at < Date.now() / 1000) return { status: 'guest' };
    return {
      status:      'authenticated',
      accessToken: data.access_token,
      sub:         data.profile?.sub ?? '',
    };
  } catch {
    return { status: 'guest' };
  }
}

export function sessionFromOidcUser(user) {
  return {
    status:      'authenticated',
    accessToken: user.access_token,
    sub:         user.profile.sub,
  };
}
