export function OrbitMark({ small = false }: { small?: boolean }) {
  return (
    <svg
      width={small ? 30 : 44}
      height={small ? 30 : 44}
      viewBox="0 0 44 44"
      fill="none"
      aria-hidden="true"
    >
      <circle cx="22" cy="22" r="9" fill="currentColor" />
      <ellipse
        cx="22"
        cy="22"
        rx="21"
        ry="10"
        transform="rotate(-35 22 22)"
        stroke="currentColor"
        strokeWidth="1.5"
      />
      <circle cx="37" cy="12" r="3" fill="currentColor" />
    </svg>
  );
}
