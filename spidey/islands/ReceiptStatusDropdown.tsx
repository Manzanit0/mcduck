import { useComputed, useSignal, effect } from "@preact/signals";
import { ReceiptStatus } from "../gen/receipts.v1/receipts_pb.ts";
import { IS_BROWSER } from "$fresh/runtime.ts";

interface ReceiptStatusDropdownProps {
  status: number
  updateStatus: (status: number) => Promise<void>;
}

export default function ReceiptStatusDropdown(
  props: ReceiptStatusDropdownProps
) {
  const open = useSignal(false);
  const statusSgn = useSignal(props.status)

  const dropdownOptions = useComputed(() => {
    const options = [pendingReview(false), reviewed(false)];
    switch (statusSgn.value) {
      case ReceiptStatus.PENDING_REVIEW:
        options[0] = pendingReview(true);
        break;
      case ReceiptStatus.REVIEWED:
        options[1] = reviewed(true);
        break;
      default:
        break;
    }

    return options;
  });

  const selectedDropdownOption = useComputed(() => {
    let option = na(false);
    switch (statusSgn.value) {
      case ReceiptStatus.PENDING_REVIEW:
        option = pendingReview(false);
        break;
      case ReceiptStatus.REVIEWED:
        option = reviewed(false);
        break;
      default:
        break;
    }
    return option;
  });

  const closeDropdown = () => (open.value = false);
  const toggleDropdown = () => (open.value = !open.value);

  // If the user clicks elsewhere outside of the dropdown, just close it.
  //
  // NOTE: we need to check if this running in the browser because the
  // "document" API is not available on the server.
  if (IS_BROWSER) {
    effect(() => {
      document.addEventListener("click", closeDropdown, true);
      return () => {
        document.removeEventListener("click", closeDropdown, true);
      };
    });
  }

  return (
    <div>
      <div class="relative mt-2">
        <button
          type="button"
          class="relative w-full cursor-default rounded-md bg-white py-1.5 pl-3 pr-10 text-left text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-gray-500 sm:text-sm sm:leading-6"
          aria-haspopup="listbox"
          aria-expanded="true"
          aria-labelledby="listbox-label"
          onClick={toggleDropdown}
        >
          {selectedDropdownOption}
          <span class="pointer-events-none absolute inset-y-0 right-0 ml-3 flex items-center pr-2">
            <svg
              class="h-5 w-5 text-gray-400"
              viewBox="0 0 20 20"
              fill="currentColor"
              aria-hidden="true"
            >
              <path
                fill-rule="evenodd"
                d="M10 3a.75.75 0 01.55.24l3.25 3.5a.75.75 0 11-1.1 1.02L10 4.852 7.3 7.76a.75.75 0 01-1.1-1.02l3.25-3.5A.75.75 0 0110 3zm-3.76 9.2a.75.75 0 011.06.04l2.7 2.908 2.7-2.908a.75.75 0 111.1 1.02l-3.25 3.5a.75.75 0 01-1.1 0l-3.25-3.5a.75.75 0 01.04-1.06z"
                clip-rule="evenodd"
              />
            </svg>
          </span>
        </button>

        <ul
          className={`absolute mt-1 max-h-56 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm ${
            open.value
              ? "z-10 opacity-100 transition ease-in duration-100"
              : "-z-10 opacity-0"
          }`}
          tabindex={-1}
          role="listbox"
          aria-labelledby="listbox-label"
          aria-activedescendant="listbox-option-3"
        >
          {dropdownOptions.value.map((x, index) => {
            const hovered = useSignal(false);
            return (
              <li
                class={`relative cursor-default select-none py-2 pl-3 pr-9 ${
                  hovered.value ? "bg-gray-100" : " text-gray-900"
                }`}
                role="option"
                onMouseEnter={() => (hovered.value = true)}
                onMouseLeave={() => (hovered.value = false)}
                onClick={async () => {
                  if (index === 0) {
                    statusSgn.value = ReceiptStatus.PENDING_REVIEW;
                  } else {
                    statusSgn.value = ReceiptStatus.REVIEWED;
                  }

                  // When the user selects and option, we can assume he wants the dropdown closed.
                  closeDropdown();

                  await props.updateStatus(statusSgn.value);
                }}
              >
                {x}
              </li>
            );
          })}
        </ul>
      </div>
    </div>
  );
}

function checkmark() {
  return (
    <span class="absolute inset-y-0 right-0 flex items-center pr-4 text-grey-900">
      <svg
        class="h-5 w-5"
        viewBox="0 0 20 20"
        fill="currentColor"
        aria-hidden="true"
      >
        <path
          fill-rule="evenodd"
          d="M16.704 4.153a.75.75 0 01.143 1.052l-8 10.5a.75.75 0 01-1.127.075l-4.5-4.5a.75.75 0 011.06-1.06l3.894 3.893 7.48-9.817a.75.75 0 011.05-.143z"
          clip-rule="evenodd"
        />
      </svg>
    </span>
  );
}

function reviewed(selected: boolean) {
  const fontClass = selected ? "font-semibold" : "font-normal";
  return (
    <>
      <div class="flex items-center">
        <div class="h-2.5 w-2.5 rounded-full bg-green-500 me-2"></div>
        <span class={`ml-3 block truncate ${fontClass}`}>Reviewed</span>
      </div>
      {selected ? checkmark() : <></>}
    </>
  );
}

function pendingReview(selected: boolean) {
  const fontClass = selected ? "font-semibold" : "font-normal";
  return (
    <>
      <div class="flex items-center">
        <div class="h-2.5 w-2.5 rounded-full bg-red-500 me-2"></div>
        <span class={`ml-3 block truncate ${fontClass}`}>Pending Review</span>
      </div>
      {selected ? checkmark() : <></>}
    </>
  );
}

function na(selected: boolean) {
  const fontClass = selected ? "font-semibold" : "font-normal";
  return (
    <>
      <div class="flex items-center">
        <div class="h-2.5 w-2.5 rounded-full bg-yellow-500 me-2"></div>
        <span class={`ml-3 block truncate ${fontClass}`}>N/a</span>
      </div>
      {selected ? checkmark() : <></>}
    </>
  );
}
