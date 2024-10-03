import Checkbox from "../components/Checkbox.tsx";
import FormattedMoney from "../components/FormattedMoney.tsx";
import GenericTable from "../components/GenericTable.tsx";
import TextInput from "../components/TextInput.tsx";
import { SerializableExpense } from "../lib/types.ts";
import { useSignal } from "@preact/signals";

interface TableProps {
  expenses: SerializableExpense[];
  url: string;
}

interface CheckeableExpense extends SerializableExpense {
  checked: boolean;
}

export default function ExpensesTable(props: TableProps) {
  const mapped = props.expenses.map((x) => {
    return useSignal<CheckeableExpense>({
      ...x,
      checked: false,
    });
  });

  const globallySelected = useSignal(false);

  const checkExpenses = () => {
    globallySelected.value = !globallySelected.value;

    for (const r of mapped) {
      r.value.checked = globallySelected.value;
    }
  };
  return (
    <div class="sm:rounded-lg">
      <GenericTable
        data={mapped}
        columns={[
          {
            header: (
              <Checkbox
                onInput={checkExpenses}
                checked={globallySelected.value}
              />
            ),
            accessor: (r) => (
              <Checkbox
                checked={r.value.checked}
                onInput={() => (r.value.checked = !r.value.checked)}
              />
            ),
          },
          {
            header: <span>Category</span>,
            accessor: (r) => (
              <TextInput
                value={r.value.category}
                onfocusout={() => (Promise.resolve())}
              />
            ),
          },
          {
            header: <span>Subcategory</span>,
            accessor: (r) => (
              <TextInput
                value={r.value.subcategory}
                onfocusout={() => (Promise.resolve())}
              />
            ),
          },
          {
            header: <span>Description</span>,
            accessor: (r) => (
              <TextInput
                value={r.value.description}
                onfocusout={() => (Promise.resolve())}
              />
            ),
          },
          {
            header: <span>Amount</span>,
            accessor: (r) => (
              <FormattedMoney
                currency="EUR"
                amount={Number(r.value.amount)}
              />
            ),
          },
          {
            header: <span>Action</span>,
            accessor: (r) => (
              <a
                href={`receipts/${r.value.id}`}
                class="font-medium text-blue-600 dark:text-blue-500 hover:underline"
              >
                Delete
              </a>
            ),
          },
        ]}
      />
    </div>
  );
}
