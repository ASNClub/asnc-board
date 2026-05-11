import { useState, useEffect } from 'react';
import {
  BrowserRouter, Routes, Route, Outlet, Navigate, useOutletContext,
  useNavigate, useParams,
} from 'react-router-dom';
import { useQueryClient } from '@tanstack/react-query';
import { AuthProvider, useAuth } from './AuthContext';
import { getMe } from './lib/api';
import { qk } from './lib/queryKeys';
import { applyAppearance, watchSystemTheme } from './lib/appearance';
import TopNav from './TopNav';
import FeedScreen from './screens/FeedScreen';
import ThreadScreen from './screens/ThreadScreen';
import ProfileScreen from './screens/ProfileScreen';
import CommunityScreen from './screens/CommunityScreen';
import CommunitiesIndexScreen from './screens/CommunitiesIndexScreen';
import CreateCommunityScreen from './screens/CreateCommunityScreen';
import PostEditorScreen from './screens/PostEditorScreen';
import SettingsScreen from './screens/SettingsScreen';
import NotificationsScreen from './screens/NotificationsScreen';
import RoadmapScreen from './screens/RoadmapScreen';
import AdminScreen from './screens/AdminScreen';
import CallbackScreen from './screens/CallbackScreen';
import LoginScreen from './screens/LoginScreen';
import OnboardingScreen from './screens/OnboardingScreen';
import { IconX } from '@tabler/icons-react';


const TWEAKS_DEFAULT = { honeycomb: true };

export const useMe = () => useOutletContext();

const AppLayout = () => {
  const { isAuthenticated } = useAuth();
  const [tweaks, setTweaks] = useState(TWEAKS_DEFAULT);
  const [tweaksVisible, setTweaksVisible] = useState(false);
  const [me, setMe] = useState(null);
  const [meLoading, setMeLoading] = useState(true);
  const navigate = useNavigate();
  const qc = useQueryClient();

  useEffect(() => {
    window.scrollTo(0, 0);
    applyAppearance();
    watchSystemTheme();
  }, []);

  useEffect(() => {
    const handler = (e) => {
      if (!e.data) return;
      if (e.data.type === '__activate_edit_mode')   setTweaksVisible(true);
      if (e.data.type === '__deactivate_edit_mode') setTweaksVisible(false);
    };
    window.addEventListener('message', handler);
    window.parent.postMessage({ type: '__edit_mode_available' }, '*');
    return () => window.removeEventListener('message', handler);
  }, []);

  useEffect(() => {
    document.body.classList.toggle('honeycomb-bg', tweaks.honeycomb);
  }, [tweaks.honeycomb]);

  useEffect(() => {
    if (!isAuthenticated) { setMe(null); setMeLoading(false); return; }
    setMeLoading(true);
    getMe()
      .then(data => { setMe(data); qc.setQueryData(qk.me(), data); })
      .catch(() => { setMe(null); qc.removeQueries({ queryKey: qk.me() }); })
      .finally(() => setMeLoading(false));
  }, [isAuthenticated]);

  const updateTweak = (key, value) => {
    setTweaks(t => {
      const next = { ...t, [key]: value };
      window.parent.postMessage({ type: '__edit_mode_set_keys', edits: { [key]: value } }, '*');
      return next;
    });
  };

  if (isAuthenticated && meLoading) {
    return (
      <div className="auth-shell">
        <div className="callback-stack">
          <div className="callback-spinner" />
          <div style={{ fontFamily: 'var(--mono)', fontSize: 12, color: 'var(--text-dim)' }}>загружаем профиль…</div>
        </div>
      </div>
    );
  }

  if (me && !me.onboardingDone) {
    return (
      <OnboardingScreen
        displayNameFromIdp={me.displayName ?? ''}
        onDone={async () => {
          try {
            const fresh = await getMe();
            setMe(fresh);
          } catch {
            setMe(m => ({ ...m, onboardingDone: true }));
          }
        }}
      />
    );
  }

  return (
    <div className="app-shell">
      <TopNav me={me} />
      <Outlet context={{ me, setMe }} />

      {tweaksVisible && (
        <div className="tweaks-panel">
          <div className="tweaks-header">
            Tweaks
            <button className="btn ghost" style={{ padding: 4 }} onClick={() => setTweaksVisible(false)}>
              <IconX size={12} stroke={2} />
            </button>
          </div>
          <div className="tweaks-body">
            <div className="tweak-row">
              <div className="tweak-label">Декор</div>
              <label className="tweak-switch">
                <input
                  type="checkbox"
                  checked={tweaks.honeycomb}
                  onChange={(e) => updateTweak('honeycomb', e.target.checked)}
                />
                <span className="sw" />
                <span className="tl">Honeycomb-фон</span>
              </label>
            </div>
            <div className="tweak-row">
              <div className="tweak-label">Быстрая навигация</div>
              <div className="tweak-segmented">
                {[
                  { p: '/',          l: 'Лента' },
                  { p: '/c',         l: 'Сообщ.' },
                  { p: '/me',        l: 'Проф.' },
                ].map(o => (
                  <button key={o.p} onClick={() => navigate(o.p)}>{o.l}</button>
                ))}
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

const RequireAuth = ({ children }) => {
  const { isAuthenticated } = useAuth();
  if (!isAuthenticated) return <Navigate to="/" replace />;
  return children;
};

const ProfileRoute = () => {
  const { username } = useParams();
  return <ProfileScreen key={username || '__me__'} username={username} />;
};

const CommunityRoute = () => {
  const { slug } = useParams();
  return <CommunityScreen key={slug} slug={slug} />;
};

const ThreadRoute = () => {
  const { postId } = useParams();
  return <ThreadScreen key={postId} postId={postId} />;
};

const PostEditorRoute = () => {
  const { slug } = useParams();
  return <PostEditorScreen slug={slug} />;
};

const PostEditRoute = () => {
  const { postId } = useParams();
  return <PostEditorScreen key={postId} postId={postId} />;
};

const App = () => (
  <AuthProvider>
    <BrowserRouter>
      <Routes>
        <Route path="/callback" element={<CallbackScreen />} />
        <Route path="/login" element={<LoginScreen />} />

        <Route element={<AppLayout />}>
          <Route index element={<HomeRoute />} />
          <Route path="/p/:postId" element={<ThreadRoute />} />
          <Route path="/c" element={<CommunitiesIndexScreen />} />
          <Route path="/c/new" element={<RequireAuth><CreateCommunityScreen /></RequireAuth>} />
          <Route path="/c/:slug" element={<CommunityRoute />} />
          <Route path="/c/:slug/submit" element={<RequireAuth><PostEditorRoute /></RequireAuth>} />
          <Route path="/submit" element={<RequireAuth><PostEditorRoute /></RequireAuth>} />
          <Route path="/p/:postId/edit" element={<RequireAuth><PostEditRoute /></RequireAuth>} />
          <Route path="/u/:username" element={<ProfileRoute />} />
          <Route path="/me" element={<ProfileRoute />} />
          <Route path="/settings" element={<RequireAuth><SettingsScreen /></RequireAuth>} />
          <Route path="/notifications" element={<RequireAuth><NotificationsScreen /></RequireAuth>} />
          <Route path="/roadmap" element={<RoadmapScreen />} />
          <Route path="/admin" element={<RequireAuth><AdminScreen /></RequireAuth>} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Route>
      </Routes>
    </BrowserRouter>
  </AuthProvider>
);

const HomeRoute = () => {
  const params = new URLSearchParams(window.location.search);
  const legacyPost = params.get('post');
  if (legacyPost) return <Navigate to={`/p/${legacyPost}`} replace />;
  return <FeedScreen />;
};

export default App;
