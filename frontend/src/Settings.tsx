import steam1img from './assets/screenshots/steam1.png'
import steam2img from './assets/screenshots/steam2.png'

import { Button, Fieldset, Group, SpaceProps, TextInput } from '@mantine/core'
import { TransformedValues, UseFormReturnType, useForm } from '@mantine/form'
import { useDisclosure } from '@mantine/hooks'
import deepEqual from 'deep-equal'
import { ReactNode, useEffect, useState } from 'react'
import { CodeHighlight } from './CodeHighlight'
import Modal from './Modal'
import cls from './Settings.module.css'
import { app } from './api'
import { Config, GameInfo, Presence } from './appstate'
import { useConfig, useLaunchOptions, usePresence } from './hooks/query'

interface FormConfig<K extends keyof Config> {
  name: K
  cfg: Config
  form: UseFormReturnType<Config[K]>
  changed: boolean
}

function useFormConfig<K extends keyof Config>(cfg: Config, name: K): FormConfig<K> {
  const form = useForm({ initialValues: cfg[name] })
  const { setValues } = form
  useEffect(() => {
    setValues(cfg[name])
  }, [setValues, cfg, name])
  return { name, cfg, form, changed: !deepEqual(form.values, cfg[name]) }
}

function Form<K extends keyof Config>({
  fc: { form, changed },
  title,
  select,
  fields,
  children,
}: {
  fc: FormConfig<K>
  title: string
  select: (m: TransformedValues<UseFormReturnType<Config[K]>>) => Partial<Config>
  fields: ReactNode
  children: ReactNode
}) {
  const [busy, setBusy] = useState(false)
  async function onSubmit(m: TransformedValues<typeof form>) {
    try {
      setBusy(true)
      await app.UpdateConfig(select(m))
    } catch (e) {
      form.setErrors({ root: `${e}` })
    } finally {
      setBusy(false)
    }
  }
  return (
    <form onSubmit={form.onSubmit(onSubmit)}>
      <Fieldset className={cls.fieldSet} legend={title} disabled={busy}>
        <Group align="flex-end">
          {fields}
          <Button color="dark" type="submit" disabled={!changed}>
            Save
          </Button>
        </Group>
        {form.errors['root'] ? <p>{form.errors['root']}</p> : null}
        {children}
      </Fieldset>
    </form>
  )
}

function NetconSteamLaunchInstructions({ launchOptions }: { launchOptions: ReactNode }) {
  return (
    <div>
      Open Steam then:
      <ol className={cls.instructionList}>
        <li>
          <span>Right click on the game in the sidebar on the left</span>
        </li>

        <li>
          <span>Click Properties on the menu that opens</span>
          <img src={steam1img} />
        </li>

        <li>
          <span>In the dialog that opens, update the launch options to include the:</span>
          <br />
          {launchOptions}
          <img src={steam2img} />
        </li>
      </ol>
    </div>
  )
}

function NetconForm({ cfg, pr, game }: { cfg: Config; pr: Presence; game: GameInfo }) {
  const [instrOpen, instr] = useDisclosure()
  const fc = useFormConfig(cfg, 'netcon')
  const lo = useLaunchOptions(pr.userID, game.id)

  if (lo.type !== 'ok') {
    return lo.alt
  }

  const { values, getInputProps } = fc.form
  const oldCmd = lo.v.replaceAll(/\s*-netcon(?:port|password)\s+\S+\s*/g, ' ').trim() || '%command%'
  const newCmd = [
    pr.gameID,
    oldCmd,
    `-netconport ${values.port}`,
    values.password ? `-netconpassword ${values.password}` : ``,
  ]
    .filter(Boolean)
    .join(' ')

  const launchOptions =
    values.port || values.password ? (
      <CodeHighlight code={newCmd} language="" copyLabel="Copy" copiedLabel="Copied!" />
    ) : null

  return (
    <div>
      <Form
        fc={fc}
        title="Netcon"
        select={({ port, password }) =>
          ({ netcon: { port: new Number(port), password } }) as Config
        }
        fields={
          <>
            <TextInput type="number" label="Port" placeholder="31173" {...getInputProps('port')} />
            <TextInput type="text" label="Password" placeholder="" {...getInputProps('password')} />
          </>
        }
      >
        {launchOptions ? (
          <>
            <br />
            <div>Steam launch options: {launchOptions}</div>

            {instrOpen ? (
              <Modal onClose={instr.close}>
                <NetconSteamLaunchInstructions launchOptions={launchOptions} />
              </Modal>
            ) : null}

            <a
              href="#"
              onClick={(e) => {
                e.preventDefault()
                instr.toggle()
              }}
            >
              {instrOpen ? 'Hide instructions' : 'How do I set launch options?'}
            </a>
          </>
        ) : null}
      </Form>
    </div>
  )
}

export interface SettingsProps {
  game: GameInfo
}

export default function Settings({ game }: SettingsProps) {
  const cfg = useConfig()
  const pr = usePresence()

  if (pr.type !== 'ok') {
    return pr.alt
  }
  if (cfg.type !== 'ok') {
    return cfg.alt
  }

  return (
    <div className={cls.root}>
      <NetconForm cfg={cfg.v} pr={pr.v} game={game} />
    </div>
  )
}
