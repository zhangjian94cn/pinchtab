const AUTH_REQUIRED_EVENT = "pinchtab-auth-required";
const AUTH_STATE_CHANGED_EVENT = "pinchtab-auth-state-changed";

export function dispatchAuthRequired(reason: string): void {
  window.dispatchEvent(
    new CustomEvent(AUTH_REQUIRED_EVENT, {
      detail: { reason },
    }),
  );
}

export function dispatchAuthStateChanged(): void {
  window.dispatchEvent(new Event(AUTH_STATE_CHANGED_EVENT));
}

export function sameOriginUrl(url: string): string {
  const absolute = new URL(url, window.location.origin);
  return absolute.pathname + absolute.search;
}

export { AUTH_REQUIRED_EVENT, AUTH_STATE_CHANGED_EVENT };
