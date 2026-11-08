import { Alert, AlertDescription } from "@/components/ui/alert";
import {
  Field,
  FieldGroup,
  FieldTitle,
  FieldDescription,
} from "@/components/ui/field";

export function SubmittedRegistrationAnswers({
  fields,
  answers,
}: {
  fields: { id: string; label: string }[];
  answers: { fieldId: string; value?: unknown }[];
}) {
  return (
    <FieldGroup>
      <Alert>
        <AlertDescription>
          Your answers have been submitted and cannot be edited for this
          application. If you leave this team or your application is rejected,
          you can submit new answers when applying again.
        </AlertDescription>
      </Alert>
      {answers.map((answer) => (
        <Field key={answer.fieldId}>
          <FieldTitle>
            {fields.find((field) => field.id === answer.fieldId)?.label ??
              answer.fieldId}
          </FieldTitle>
          <FieldDescription className="whitespace-pre-wrap">
            {Array.isArray(answer.value)
              ? answer.value.map(String).join(", ") || "No answer"
              : String(answer.value ?? "") || "No answer"}
          </FieldDescription>
        </Field>
      ))}
    </FieldGroup>
  );
}
