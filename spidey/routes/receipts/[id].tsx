import { RouteContext } from "$fresh/server.ts";
import ExpensesTable from "../../islands/ExpensesTable.tsx";
import { createPromiseClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { ReceiptsService } from "../../gen/receipts.v1/receipts_connect.ts";
import { AuthState } from "../../lib/auth.ts";
import { mapExpensesToSerializable } from "../../lib/types.ts";

const url = Deno.env.get("API_HOST")!;

export default async function Single(
  _req: Request,
  ctx: RouteContext<unknown, AuthState>,
) {
  console.log("get receipt");
  const transport = createConnectTransport({
    baseUrl: url!,
  });

  const client = createPromiseClient(ReceiptsService, transport);

  const res = await client.getReceipt({ id: BigInt(ctx.params.id) }, {
    headers: { authorization: `Bearer ${ctx.state.authToken}` },
  });

  return (
    <div class="m-6">
      <h1>Receipt {ctx.params.id}</h1>
      <ExpensesTable
        expenses={mapExpensesToSerializable(res.receipt!.expenses)}
        url={url}
      />
    </div>
  );
}
