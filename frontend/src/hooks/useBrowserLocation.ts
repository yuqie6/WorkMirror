import { useCallback, useEffect, useState } from 'react';

export interface BrowserLocationSnapshot {
  pathname: string;
  search: string;
  hash: string;
}

function readSnapshot(): BrowserLocationSnapshot {
  return {
    pathname: window.location.pathname || '/',
    search: window.location.search || '',
    hash: window.location.hash || '',
  };
}

export function useBrowserLocation() {
  const [location, setLocation] = useState<BrowserLocationSnapshot>(() => readSnapshot());

  useEffect(() => {
    const onPopState = () => setLocation(readSnapshot());
    window.addEventListener('popstate', onPopState);
    return () => window.removeEventListener('popstate', onPopState);
  }, []);

  const navigate = useCallback((to: string, opts?: { replace?: boolean }) => {
    const target = typeof to === 'string' && to.trim() !== '' ? to : '/';
    if (opts?.replace) {
      window.history.replaceState(null, '', target);
    } else {
      window.history.pushState(null, '', target);
    }
    setLocation(readSnapshot());
  }, []);

  return { location, navigate };
}

