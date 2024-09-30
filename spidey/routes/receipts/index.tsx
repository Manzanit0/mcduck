import { RouteContext } from "$fresh/server.ts";
import { createPromiseClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { ReceiptsService } from "../../gen/receipts.v1/receipts_connect.ts";
import {
  ListReceiptsSince,
  Receipt,
} from "../../gen/receipts.v1/receipts_pb.ts";
import { State } from "../_middleware.ts";

const url = Deno.env.get("API_HOST")!;

export default async function List(_req: Request, ctx: RouteContext<State>) {
  console.log("list receipts");
  if (!ctx.state || !ctx.state.loggedIn) {
    return <div>state: {JSON.stringify(ctx.state)}</div>;
  }

  const transport = createConnectTransport({
    baseUrl: url!,
  });

  const client = createPromiseClient(ReceiptsService, transport);

  const res = await client.listReceipts(
    { since: ListReceiptsSince.ALL_TIME },
    { headers: { authorization: `Bearer ${ctx.state.authToken}` } },
  );

  return (
    <div class="relative overflow-x-auto">
      <table class="w-full text-sm text-left rtl:text-right text-gray-500 dark:text-gray-400">
        <thead class="text-xs text-gray-700 uppercase bg-gray-50 dark:bg-gray-700 dark:text-gray-400">
          <tr>
            <th scope="col" class="px-6 py-3">
              Date
            </th>
            <th scope="col" class="px-6 py-3">
              Vendor
            </th>
            <th scope="col" class="px-6 py-3">
              Total Amount
            </th>
            <th scope="col" class="px-6 py-3">
              Status
            </th>
          </tr>
        </thead>
        <tbody>
          {res.receipts.map((r) => row(r))}
        </tbody>
      </table>
    </div>
  );
}

function row(r: Receipt) {
  const total = r.expenses.reduce((acc, ex) => acc += ex.amount, BigInt(0));
  const totalFmt = new Intl.NumberFormat("de-DE", {
    style: "currency",
    currency: "EUR",
  }).format(Number(total)/100);

  return (
    <tr class="bg-white border-b dark:bg-gray-800 dark:border-gray-700">
      <td class="px-6 py-4">{r.date?.toDate().toDateString()}</td>
      <td class="px-6 py-4">{r.vendor}</td>
      <td class="px-6 py-4">{totalFmt}</td>
      <td class="px-6 py-4">{r.status == 1 ? "Pending Review" : "Reviewed"}</td>
    </tr>
  );
}
