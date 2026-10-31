import { SettingsRounded, TuneRounded } from '@mui/icons-material'
import {
  Box,
  Collapse,
  MenuItem,
  Select,
  TextField,
  Typography,
} from '@mui/material'
import React from 'react'
import { useTranslation } from 'react-i18next'

import { Switch } from '@/components/base'
import { useVerge } from '@/hooks/use-verge'
import { mutate } from '@/services/mutate'
import { showNotice } from '@/services/notice-service'

interface Props {
  onError?: (err: Error) => void
}

export const FragmentCard: React.FC<Props> = ({ onError }) => {
  const { t } = useTranslation()
  const { verge, patchVerge, mutateVerge } = useVerge()

  const enabled = verge?.enable_fragment ?? false
  const packets = verge?.fragment_packets ?? 'tlshello'
  const length = verge?.fragment_length ?? '100-200'
  const interval = verge?.fragment_interval ?? '10-20'

  const handleToggle = async (val: boolean) => {
    try {
      mutateVerge({ ...verge, enable_fragment: val }, false)
      await patchVerge({ enable_fragment: val })
      showNotice.success(
        val
          ? 'Фрагментирование включено (Обход DPI активен)'
          : 'Фрагментирование выключено',
        1500,
      )
    } catch (err: any) {
      onError?.(err)
      showNotice.error(err)
    }
  }

  const handlePacketsChange = async (val: string) => {
    try {
      mutateVerge({ ...verge, fragment_packets: val }, false)
      await patchVerge({ fragment_packets: val })
    } catch (err: any) {
      onError?.(err)
    }
  }

  const handleLengthChange = async (val: string) => {
    try {
      mutateVerge({ ...verge, fragment_length: val }, false)
      await patchVerge({ fragment_length: val })
    } catch (err: any) {
      onError?.(err)
    }
  }

  const handleIntervalChange = async (val: string) => {
    try {
      mutateVerge({ ...verge, fragment_interval: val }, false)
      await patchVerge({ fragment_interval: val })
    } catch (err: any) {
      onError?.(err)
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
              {t('settings.sections.clash.form.fields.fragment.title', 'Фрагментирование')}
            </Typography>
            <Typography
              variant="caption"
              sx={{ color: 'text.secondary', display: 'block', mt: 0.3 }}
            >
              {t(
                'settings.sections.clash.form.fields.fragment.subtitle',
                'Обход DPI через фрагментацию пакетов',
              )}
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
          {/* Row 1: Packets */}
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
                {t('settings.sections.clash.form.fields.fragment.packets', 'Packets')}
              </Typography>
            </Box>

            <Select
              size="small"
              value={packets}
              onChange={(e) => void handlePacketsChange(e.target.value)}
              sx={{
                width: 140,
                borderRadius: '8px',
                '> div': { py: '6px' },
              }}
            >
              <MenuItem value="tlshello">tlshello</MenuItem>
              <MenuItem value="1-1">1-1</MenuItem>
              <MenuItem value="1-2">1-2</MenuItem>
              <MenuItem value="1-3">1-3</MenuItem>
            </Select>
          </Box>

          {/* Row 2: Length */}
          <Box
            sx={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
            }}
          >
            <Typography variant="body2" sx={{ fontWeight: 500, pl: 5.5 }}>
              {t('settings.sections.clash.form.fields.fragment.length', 'Length')}
            </Typography>
            <TextField
              size="small"
              value={length}
              onChange={(e) => void handleLengthChange(e.target.value)}
              placeholder="100-200"
              sx={{
                width: 140,
                '& .MuiOutlinedInput-root': { borderRadius: '8px' },
                input: { py: '6px', textAlign: 'center' },
              }}
            />
          </Box>

          {/* Row 3: Interval */}
          <Box
            sx={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
            }}
          >
            <Typography variant="body2" sx={{ fontWeight: 500, pl: 5.5 }}>
              {t('settings.sections.clash.form.fields.fragment.interval', 'Interval')}
            </Typography>
            <TextField
              size="small"
              value={interval}
              onChange={(e) => void handleIntervalChange(e.target.value)}
              placeholder="10-20"
              sx={{
                width: 140,
                '& .MuiOutlinedInput-root': { borderRadius: '8px' },
                input: { py: '6px', textAlign: 'center' },
              }}
            />
          </Box>
        </Box>
      </Collapse>
    </Box>
  )
}
