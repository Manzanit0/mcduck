import { Receipt } from "../gen/receipts.v1/receipts_pb.ts";

export interface SerializableReceipt {
    id: bigint;
    status: string;
    vendor: string;
    date?: string;
    expenses: SerializableExpense[];
}

export interface SerializableExpense {
    id: bigint;
    date?: string;
    category: string;
    subcategory: string;
    description: string;
    amount: bigint;
}

export function mapReceiptsToSerializable(receipts: Receipt[]): SerializableReceipt[] {
    return receipts.map((r) => {
        return {
            id: r.id,
            status: r.status.toString(),
            vendor: r.vendor,
            date: r.date?.toDate().toISOString(),
            expenses: r.expenses.map((e) => {
                return {
                    id: e.id,
                    date: e.date?.toDate().toISOString(),
                    category: e.category,
                    subcategory: e.subcategory,
                    description: e.description,
                    amount: e.amount,
                };
            }),
        };
    });
}
