import { useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { IconCheck } from '@tabler/icons-react';
import { userManager } from '../lib/auth';

const STEPS = [
  'Получили токен',
  'Проверили подпись',
  'Провизионируем профиль',
  'Готовим ленту',
];

const CallbackScreen = () => {
  const called = useRef(false);
  const navigate = useNavigate();
  const [phase, setPhase] = useState(0);

  useEffect(() => {
    if (called.current) return;
    called.current = true;

    if (!userManager) {
      navigate('/', { replace: true });
      return;
    }

    setPhase(1);
    userManager
      .signinRedirectCallback()
      .then(() => {
        setPhase(3);
        setTimeout(() => navigate('/', { replace: true }), 250);
      })
      .catch(() => navigate('/', { replace: true }));
  }, [navigate]);

  return (
    <div className="auth-shell">
      <div className="auth-card">
        <div className="auth-head">
          <div className="auth-hex">H</div>
          <h1 className="auth-title">Входим в <span className="amp">сад</span>…</h1>
          <p className="auth-sub">Подключение к Zitadel занимает пару секунд</p>
        </div>

        <div className="auth-body">
          <div className="callback-stack">
            <div className="callback-spinner" />
            <div className="callback-steps">
              {STEPS.map((label, i) => {
                const state = i < phase ? 'done' : i === phase ? 'cur' : '';
                return (
                  <div key={i} className={'callback-step ' + state}>
                    <span className="mark">
                      {state === 'done' && <IconCheck size={14} stroke={2.5} />}
                      {state === 'cur'  && <span className="pulse" />}
                    </span>
                    {label}
                  </div>
                );
              })}
            </div>
          </div>
        </div>

        <div className="auth-foot">
          не редиректит? <a onClick={() => navigate('/login', { replace: true })}>назад на /login</a>
        </div>
      </div>
    </div>
  );
};

export default CallbackScreen;
