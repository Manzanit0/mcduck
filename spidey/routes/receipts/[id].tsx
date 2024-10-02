import { PageProps } from "$fresh/server.ts";

export default function Greet(props: PageProps) {
  return <div>Receipt {props.params.id}</div>;
}
