import { RouteContext } from "$fresh/server.ts";
import { createPromiseClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { ReceiptsService } from "../../gen/receipts.v1/receipts_connect.ts";
import { ListReceiptsSince } from "../../gen/receipts.v1/receipts_pb.ts";
import ReceiptsTable from "../../islands/ReceiptsTable.tsx";
import { mapReceiptsToSerializable } from "../../lib/types.ts";
import { AuthState } from "../../lib/auth.ts";

const url = Deno.env.get("API_HOST")!;

export default async function List(
  _req: Request,
  ctx: RouteContext<unknown, AuthState>,
) {
  console.log("list receipts");
  if (!ctx.state || !ctx.state.loggedIn) {
    return ctx.renderNotFound({});
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
    <div class="m-6">
      <ReceiptsTable
        receipts={mapReceiptsToSerializable(res.receipts)}
        url={url}
      />
    </div>
  );
}
