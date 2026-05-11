import { IconBrandGithub, IconCircleCheck } from '@tabler/icons-react';
import { useAuth } from '../AuthContext';

const LoginScreen = () => {
  const { login, loginWith } = useAuth();

  return (
    <div className="auth-shell">
      <div className="auth-card">
        <div className="auth-head">
          <div className="auth-hex">H</div>
          <h1 className="auth-title">Добро пожаловать в <span className="amp">сад</span></h1>
          <p className="auth-sub">Единый аккаунт через Zitadel — для форума, git и wakapi. Email-верификация обязательна.</p>
        </div>

        <div className="auth-body">
          <button className="auth-btn-primary" onClick={login}>
            <IconCircleCheck size={15} stroke={2} />
            Войти через Zitadel
          </button>

          <div className="auth-divider">или через GitHub</div>

          <button className="auth-btn-ghost" onClick={() => loginWith('github')}>
            <IconBrandGithub size={16} stroke={2} />
            GitHub
          </button>
        </div>

        <div className="auth-foot">
          Создавая аккаунт, ты соглашаешься с правилами и политикой конфиденциальности
        </div>
      </div>
    </div>
  );
};

export default LoginScreen;
