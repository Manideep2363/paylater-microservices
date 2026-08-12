import { useId, useState, type InputHTMLAttributes } from 'react'
import './PasswordInput.css'

type PasswordInputProps = Omit<
  InputHTMLAttributes<HTMLInputElement>,
  'type'
> & {
  label: string
}

function PasswordInput({
  label,
  id,
  className,
  disabled,
  ...inputProps
}: PasswordInputProps) {
  const generatedId = useId()
  const inputId = id ?? generatedId
  const [visible, setVisible] = useState(false)

  return (
    <div className={className ? `password-field ${className}` : 'password-field'}>
      <label htmlFor={inputId}>{label}</label>
      <div className="password-field__control">
        <input
          {...inputProps}
          id={inputId}
          type={visible ? 'text' : 'password'}
          disabled={disabled}
          className="password-field__input"
        />
        <button
          type="button"
          className="password-field__toggle"
          onClick={() => setVisible((value) => !value)}
          aria-label={visible ? 'Hide password' : 'Show password'}
          aria-pressed={visible}
          disabled={disabled}
        >
          <span aria-hidden="true">{visible ? 'Hide' : 'Show'}</span>
        </button>
      </div>
    </div>
  )
}

export default PasswordInput
