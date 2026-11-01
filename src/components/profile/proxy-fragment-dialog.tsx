// Copyright (C) 2026 mateika8814 (https://github.com/mateika8814)
// This file is part of allah-verge (fork of Clash Verge Rev).
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

import { TuneRounded } from '@mui/icons-material'
import {
  Alert,
  Box,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormHelperText,
  MenuItem,
  Select,
  TextField,
  Typography,
} from '@mui/material'
import React, { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Switch } from '@/components/base'
import { isTlsProxy, validateIntervalRange, validateLengthRange } from '@/utils/fragment'

interface Props {
  open: boolean
  proxy: IProxyConfig | null
  onClose: () => void
  onSave: (updated: IProxyConfig) => void
}

interface InnerProps {
  proxy: IProxyConfig
  onClose: () => void
  onSave: (updated: IProxyConfig) => void
}

const ProxyFragmentDialogInner: React.FC<InnerProps> = ({
  proxy,
  onClose,
  onSave,
}) => {
  const { t } = useTranslation()

  const frag = proxy.fragment
  const [enabled, setEnabled] = useState(frag?.enabled ?? false)
  const [mode, setMode] = useState<
    'sni' | 'random' | 'record_fragment' | 'both'
  >((frag?.mode as any) ?? 'sni')
  const [length, setLength] = useState(frag?.length ?? '10-30')
  const [intervalVal, setIntervalVal] = useState(frag?.interval ?? '1-5')
  const [recordFragment, setRecordFragment] = useState(
    frag?.['record-fragment'] ?? false,
  )

  const [lengthError, setLengthError] = useState(false)
  const [intervalError, setIntervalError] = useState(false)

  const isTls = isTlsProxy(proxy)

  const handleSave = () => {
    if (!isTls) {
      onClose()
      return
    }

    if (enabled) {
      if ((mode === 'random' || mode === 'both') && !validateLengthRange(length)) {
        setLengthError(true)
        return
      }
      if (!validateIntervalRange(intervalVal)) {
        setIntervalError(true)
        return
      }
    }

    const updatedProxy: IProxyConfig = { ...proxy }

    if (!enabled) {
      delete updatedProxy.fragment
    } else {
      updatedProxy.fragment = {
        enabled: true,
        mode,
        ...(mode === 'random' || mode === 'both' ? { length: length.trim() } : {}),
        interval: intervalVal.trim(),
        'record-fragment': recordFragment,
      }
    }

    onSave(updatedProxy)
    onClose()
  }

  return (
    <>
      <DialogTitle sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
        <TuneRounded color="primary" />
        <Box sx={{ overflow: 'hidden' }}>
          <Typography variant="h6" sx={{ fontSize: 16, fontWeight: 700 }} noWrap>
            {t('profiles.modals.proxiesEditor.fragmentDialog.title')}
          </Typography>
          <Typography
            variant="caption"
            sx={{ color: 'text.secondary', display: 'block' }}
            noWrap
          >
            {proxy.name} ({proxy.type})
          </Typography>
        </Box>
      </DialogTitle>

      <DialogContent dividers>
        {!isTls ? (
          <Alert severity="info" sx={{ mt: 1 }}>
            {t('profiles.modals.proxiesEditor.fragmentDialog.notTls')}
          </Alert>
        ) : (
          <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, pt: 1 }}>
            {/* Enable switch */}
            <Box
              sx={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
              }}
            >
              <Box>
                <Typography variant="body2" sx={{ fontWeight: 600 }}>
                  {t('profiles.modals.proxiesEditor.fragmentDialog.enabled')}
                </Typography>
                <Typography variant="caption" sx={{ color: 'text.secondary' }}>
                  {t('profiles.modals.proxiesEditor.fragmentDialog.subtitle')}
                </Typography>
              </Box>
              <Switch checked={enabled} onChange={(_, val) => setEnabled(val)} />
            </Box>

            {enabled && (
              <>
                {/* Mode Select */}
                <Box
                  sx={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                  }}
                >
                  <Typography variant="body2" sx={{ fontWeight: 500 }}>
                    {t('profiles.modals.proxiesEditor.fragmentDialog.mode')}
                  </Typography>
                  <Select
                    size="small"
                    value={mode}
                    onChange={(e) => setMode(e.target.value as any)}
                    sx={{ width: 160 }}
                  >
                    <MenuItem value="sni">
                      {t('profiles.modals.proxiesEditor.fragmentDialog.modeSni')}
                    </MenuItem>
                    <MenuItem value="random">
                      {t('profiles.modals.proxiesEditor.fragmentDialog.modeRandom')}
                    </MenuItem>
                    <MenuItem value="record_fragment">Record Fragment</MenuItem>
                    <MenuItem value="both">Both (Record + TCP)</MenuItem>
                  </Select>
                </Box>

                {/* Length (when mode === 'random' or mode === 'both') */}
                {(mode === 'random' || mode === 'both') && (
                  <Box
                    sx={{
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'space-between',
                    }}
                  >
                    <Box>
                      <Typography variant="body2" sx={{ fontWeight: 500 }}>
                        {t('profiles.modals.proxiesEditor.fragmentDialog.length')}
                      </Typography>
                      {lengthError && (
                        <FormHelperText error sx={{ m: 0 }}>
                          {t(
                            'profiles.modals.proxiesEditor.fragmentDialog.errors.invalidLength',
                          )}
                        </FormHelperText>
                      )}
                    </Box>
                    <TextField
                      size="small"
                      value={length}
                      error={lengthError}
                      placeholder="10-30"
                      onChange={(e) => {
                        setLength(e.target.value)
                        if (lengthError) setLengthError(false)
                      }}
                      sx={{ width: 140, input: { textAlign: 'center' } }}
                    />
                  </Box>
                )}

                {/* Interval */}
                <Box
                  sx={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                  }}
                >
                  <Box>
                    <Typography variant="body2" sx={{ fontWeight: 500 }}>
                      {t('profiles.modals.proxiesEditor.fragmentDialog.interval')}
                    </Typography>
                    {intervalError && (
                      <FormHelperText error sx={{ m: 0 }}>
                        {t(
                          'profiles.modals.proxiesEditor.fragmentDialog.errors.invalidInterval',
                        )}
                      </FormHelperText>
                    )}
                  </Box>
                  <TextField
                    size="small"
                    value={intervalVal}
                    error={intervalError}
                    placeholder="1-5"
                    onChange={(e) => {
                      setIntervalVal(e.target.value)
                      if (intervalError) setIntervalError(false)
                    }}
                    sx={{ width: 140, input: { textAlign: 'center' } }}
                  />
                </Box>

                {/* Record Fragment switch */}
                <Box
                  sx={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                  }}
                >
                  <Box>
                    <Typography variant="body2" sx={{ fontWeight: 500 }}>
                      {t(
                        'profiles.modals.proxiesEditor.fragmentDialog.recordFragment',
                      )}
                    </Typography>
                    <Typography
                      variant="caption"
                      sx={{ color: 'text.secondary' }}
                    >
                      {t(
                        'profiles.modals.proxiesEditor.fragmentDialog.recordFragmentDesc',
                      )}
                    </Typography>
                  </Box>
                  <Switch
                    checked={recordFragment}
                    onChange={(_, val) => setRecordFragment(val)}
                  />
                </Box>
              </>
            )}
          </Box>
        )}
      </DialogContent>

      <DialogActions>
        <Button onClick={onClose}>
          {t('profiles.modals.proxiesEditor.fragmentDialog.cancel')}
        </Button>
        <Button variant="contained" onClick={handleSave}>
          {t('profiles.modals.proxiesEditor.fragmentDialog.save')}
        </Button>
      </DialogActions>
    </>
  )
}

export const ProxyFragmentDialog: React.FC<Props> = ({
  open,
  proxy,
  onClose,
  onSave,
}) => {
  if (!proxy) return null

  return (
    <Dialog open={open} onClose={onClose} maxWidth="xs" fullWidth>
      <ProxyFragmentDialogInner
        key={`${proxy.name}-${proxy.fragment?.enabled ? '1' : '0'}`}
        proxy={proxy}
        onClose={onClose}
        onSave={onSave}
      />
    </Dialog>
  )
}
