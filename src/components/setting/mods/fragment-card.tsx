import { SettingsRounded, TuneRounded } from '@mui/icons-material'
import {
  Box,
  Collapse,
  FormHelperText,
  MenuItem,
  Select,
  TextField,
  Typography,
} from '@mui/material'
import React, { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Switch } from '@/components/base'
import { useVerge } from '@/hooks/use-verge'
import { showNotice } from '@/services/notice-service'
import { validateIntervalRange, validateLengthRange } from '@/utils/fragment'

interface Props {
  onError?: (err: Error) => void
}

export const FragmentCard: React.FC<Props> = ({ onError }) => {
  const { t } = useTranslation()
  const { verge, patchVerge, mutateVerge } = useVerge()

  const enabled = verge?.enable_fragment ?? false
  const mode = verge?.fragment_mode ?? 'sni'
  const length = verge?.fragment_length ?? '10-30'
  const interval = verge?.fragment_interval ?? '1-5'
  const recordFragment = verge?.fragment_record_fragment ?? false
  const applyAll = verge?.fragment_apply_all ?? true

  const [draftLength, setDraftLength] = useState<string | null>(null)
  const [draftInterval, setDraftInterval] = useState<string | null>(null)
  const [lengthError, setLengthError] = useState(false)
  const [intervalError, setIntervalError] = useState(false)

  const lengthDisplay = draftLength ?? length
  const intervalDisplay = draftInterval ?? interval

  const handleToggle = async (val: boolean) => {
    try {
      mutateVerge({ ...verge, enable_fragment: val }, false)
      await patchVerge({ enable_fragment: val })
      showNotice.success(
        val
          ? t('settings.sections.clash.form.fields.fragment.noticeEnabled')
          : t('settings.sections.clash.form.fields.fragment.noticeDisabled'),
        1500,
      )
    } catch (err: any) {
      onError?.(err)
      showNotice.error(err)
    }
  }

  const handleModeChange = async (newMode: string) => {
    try {
      mutateVerge({ ...verge, fragment_mode: newMode }, false)
      await patchVerge({ fragment_mode: newMode })
    } catch (err: any) {
      onError?.(err)
      showNotice.error(err)
    }
  }

  const handleLengthBlur = async () => {
    const val = lengthDisplay.trim()
    if (!validateLengthRange(val)) {
      setLengthError(true)
      return
    }
    setLengthError(false)
    setDraftLength(null)
    try {
      mutateVerge({ ...verge, fragment_length: val }, false)
      await patchVerge({ fragment_length: val })
    } catch (err: any) {
      onError?.(err)
      showNotice.error(err)
    }
  }

  const handleIntervalBlur = async () => {
    const val = intervalDisplay.trim()
    if (!validateIntervalRange(val)) {
      setIntervalError(true)
      return
    }
    setIntervalError(false)
    setDraftInterval(null)
    try {
      mutateVerge({ ...verge, fragment_interval: val }, false)
      await patchVerge({ fragment_interval: val })
    } catch (err: any) {
      onError?.(err)
      showNotice.error(err)
    }
  }

  const handleRecordFragmentToggle = async (val: boolean) => {
    try {
      mutateVerge({ ...verge, fragment_record_fragment: val }, false)
      await patchVerge({ fragment_record_fragment: val })
    } catch (err: any) {
      onError?.(err)
      showNotice.error(err)
    }
  }

  const handleApplyAllToggle = async (val: boolean) => {
    try {
      mutateVerge({ ...verge, fragment_apply_all: val }, false)
      await patchVerge({ fragment_apply_all: val })
    } catch (err: any) {
      onError?.(err)
      showNotice.error(err)
    }
  }

  return (
    <Box
      sx={{
        p: 1.5,
        mb: 1.5,
        borderRadius: '14px',
        bgcolor: (theme) =>
          theme.palette.mode === 'dark'
            ? 'rgba(255, 255, 255, 0.04)'
            : 'rgba(0, 0, 0, 0.03)',
        border: '1px solid',
        borderColor: (theme) =>
          theme.palette.mode === 'dark'
            ? 'rgba(255, 255, 255, 0.08)'
            : 'rgba(0, 0, 0, 0.08)',
      }}
    >
      {/* Top Header Row with Switch */}
      <Box
        sx={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
        }}
      >
        <Box sx={{ display: 'flex', alignItems: 'center' }}>
          <Box
            sx={{
              width: 38,
              height: 38,
              borderRadius: '10px',
              bgcolor: (theme) =>
                theme.palette.mode === 'dark'
                  ? 'rgba(46, 125, 50, 0.28)'
                  : 'rgba(76, 175, 80, 0.15)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              color: '#66bb6a',
              mr: 1.5,
              flexShrink: 0,
            }}
          >
            <TuneRounded fontSize="small" />
          </Box>
          <Box>
            <Typography variant="body1" sx={{ fontWeight: 600, lineHeight: 1.2 }}>
              {t('settings.sections.clash.form.fields.fragment.title')}
            </Typography>
            <Typography
              variant="caption"
              sx={{ color: 'text.secondary', display: 'block', mt: 0.3 }}
            >
              {t('settings.sections.clash.form.fields.fragment.subtitle')}
            </Typography>
          </Box>
        </Box>

        <Switch
          edge="end"
          checked={enabled}
          onChange={(_, checked) => void handleToggle(checked)}
        />
      </Box>

      {/* Expanded Controls when Enabled */}
      <Collapse in={enabled} timeout="auto" unmountOnExit>
        <Box
          sx={{
            mt: 2,
            pt: 1.5,
            borderTop: '1px solid',
            borderColor: (theme) =>
              theme.palette.mode === 'dark'
                ? 'rgba(255, 255, 255, 0.06)'
                : 'rgba(0, 0, 0, 0.06)',
            display: 'flex',
            flexDirection: 'column',
            gap: 1.5,
          }}
        >
          {/* Row 1: Mode Selection (sni / random) */}
          <Box
            sx={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
            }}
          >
            <Box sx={{ display: 'flex', alignItems: 'center' }}>
              <Box
                sx={{
                  width: 32,
                  height: 32,
                  borderRadius: '8px',
                  bgcolor: (theme) =>
                    theme.palette.mode === 'dark'
                      ? 'rgba(46, 125, 50, 0.28)'
                      : 'rgba(76, 175, 80, 0.15)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  color: '#66bb6a',
                  mr: 1.5,
                  flexShrink: 0,
                }}
              >
                <SettingsRounded sx={{ fontSize: 18 }} />
              </Box>
              <Typography variant="body2" sx={{ fontWeight: 500 }}>
                {t('settings.sections.clash.form.fields.fragment.mode')}
              </Typography>
            </Box>

            <Select
              size="small"
              value={mode}
              onChange={(e) => void handleModeChange(e.target.value)}
              sx={{
                width: 170,
                borderRadius: '8px',
                '> div': { py: '6px' },
              }}
            >
              <MenuItem value="sni">
                {t('settings.sections.clash.form.fields.fragment.modeSni')}
              </MenuItem>
              <MenuItem value="random">
                {t('settings.sections.clash.form.fields.fragment.modeRandom')}
              </MenuItem>
            </Select>
          </Box>

          {/* Row 2: Length (Visible ONLY when mode === 'random') */}
          {mode === 'random' && (
            <Box
              sx={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
              }}
            >
              <Box sx={{ pl: 5.5 }}>
                <Typography variant="body2" sx={{ fontWeight: 500 }}>
                  {t('settings.sections.clash.form.fields.fragment.length')}
                </Typography>
                {lengthError && (
                  <FormHelperText error sx={{ m: 0 }}>
                    {t('settings.sections.clash.form.fields.fragment.errors.invalidLength')}
                  </FormHelperText>
                )}
              </Box>
              <TextField
                size="small"
                value={lengthDisplay}
                error={lengthError}
                onChange={(e) => {
                  setDraftLength(e.target.value)
                  if (lengthError) setLengthError(false)
                }}
                onBlur={handleLengthBlur}
                placeholder="10-30"
                sx={{
                  width: 140,
                  '& .MuiOutlinedInput-root': { borderRadius: '8px' },
                  input: { py: '6px', textAlign: 'center' },
                }}
              />
            </Box>
          )}

          {/* Row 3: Interval */}
          <Box
            sx={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
            }}
          >
            <Box sx={{ pl: 5.5 }}>
              <Typography variant="body2" sx={{ fontWeight: 500 }}>
                {t('settings.sections.clash.form.fields.fragment.interval')}
              </Typography>
              {intervalError && (
                <FormHelperText error sx={{ m: 0 }}>
                  {t('settings.sections.clash.form.fields.fragment.errors.invalidInterval')}
                </FormHelperText>
              )}
            </Box>
            <TextField
              size="small"
              value={intervalDisplay}
              error={intervalError}
              onChange={(e) => {
                setDraftInterval(e.target.value)
                if (intervalError) setIntervalError(false)
              }}
              onBlur={handleIntervalBlur}
              placeholder="1-5"
              sx={{
                width: 140,
                '& .MuiOutlinedInput-root': { borderRadius: '8px' },
                input: { py: '6px', textAlign: 'center' },
              }}
            />
          </Box>

          {/* Row 4: Record Fragment Toggle */}
          <Box
            sx={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              pl: 5.5,
            }}
          >
            <Box>
              <Typography variant="body2" sx={{ fontWeight: 500 }}>
                {t('settings.sections.clash.form.fields.fragment.recordFragment')}
              </Typography>
              <Typography
                variant="caption"
                sx={{ color: 'text.secondary', display: 'block' }}
              >
                {t('settings.sections.clash.form.fields.fragment.recordFragmentDesc')}
              </Typography>
            </Box>
            <Switch
              edge="end"
              checked={recordFragment}
              onChange={(_, checked) => void handleRecordFragmentToggle(checked)}
            />
          </Box>

          {/* Row 5: Apply to all nodes Toggle */}
          <Box
            sx={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              pl: 5.5,
            }}
          >
            <Box>
              <Typography variant="body2" sx={{ fontWeight: 500 }}>
                {t('settings.sections.clash.form.fields.fragment.applyAll')}
              </Typography>
              <Typography
                variant="caption"
                sx={{ color: 'text.secondary', display: 'block' }}
              >
                {t('settings.sections.clash.form.fields.fragment.applyAllDesc')}
              </Typography>
            </Box>
            <Switch
              edge="end"
              checked={applyAll}
              onChange={(_, checked) => void handleApplyAllToggle(checked)}
            />
          </Box>
        </Box>
      </Collapse>
    </Box>
  )
}
