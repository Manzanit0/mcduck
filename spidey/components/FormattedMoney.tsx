type Props = {
  amount: number;
  currency: string;
};

export default function FormattedMoney({ amount, currency }: Props) {
  const fmt = new Intl.NumberFormat("de-DE", {
    style: "currency",
    currency: currency,
  }).format(amount / 100);
  return <span>{fmt}</span>;
}
